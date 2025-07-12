package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/orca-ng/orca/internal/config"
	"github.com/orca-ng/orca/internal/crypto"
	"github.com/orca-ng/orca/internal/database"
	"github.com/orca-ng/orca/internal/handlers"
	"github.com/orca-ng/orca/internal/middleware"
	"github.com/orca-ng/orca/internal/pipeline"
	pipelinehandlers "github.com/orca-ng/orca/internal/pipeline/handlers"
	"github.com/orca-ng/orca/internal/services"
	"github.com/sirupsen/logrus"
)

// Static files embedding is handled in embed.go and embed_dev.go

func main() {
	// Parse command line flags
	var (
		resetPassword = flag.String("reset-password", "", "Reset admin password (provide new password)")
		username      = flag.String("username", "admin", "Username for password reset (default: admin)")
	)
	flag.Parse()

	// Handle password reset mode
	if *resetPassword != "" {
		if err := handlePasswordReset(*username, *resetPassword); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully reset password for user: %s\n", *username)
		os.Exit(0)
	}
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	// Configure logging
	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		logrus.WithError(err).Warn("Invalid log level, defaulting to info")
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to database using GORM
	dbConfig := database.DatabaseConfig{
		Driver: config.GetDatabaseDriver(),
		DSN:    cfg.Database.URL,
	}
	
	db, err := database.NewGormConnection(dbConfig)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Run auto-migration
	if err := db.AutoMigrate(); err != nil {
		logrus.WithError(err).Fatal("Failed to run database migrations")
	}

	// Seed default data
	if err := db.SeedDefaultData(); err != nil {
		logrus.WithError(err).Fatal("Failed to seed default data")
	}

	// Set Gin mode
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:          12 * time.Hour,
	}))

	// Session timeout
	sessionTimeout := time.Duration(cfg.Session.SessionTimeout) * time.Minute
	logrus.WithFields(logrus.Fields{
		"session_timeout_minutes": cfg.Session.SessionTimeout,
		"session_timeout": sessionTimeout,
	}).Info("Session timeout configured")

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, sessionTimeout)
	
	// Get encryption key
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		if os.Getenv("APP_ENV") != "development" {
			logrus.Fatal("ENCRYPTION_KEY environment variable must be set in production")
		}
		// Use a default key for development only
		encryptionKey = "development-key-do-not-use-in-prod-32bytes!!"
	}
	
	// Initialize certificate manager
	certManager := services.NewCertificateManager(db, logrus.StandardLogger())
	
	// Initialize operation event service
	eventService := services.NewOperationEventService(logrus.StandardLogger())
	
	// Initialize sync job service
	syncJobService := services.NewSyncJobService(db, logrus.StandardLogger(), eventService)
	
	// Initialize pipeline processor
	pipelineConfig := &pipeline.PipelineConfig{
		TotalCapacity:      1, // Process one operation at a time
		DefaultTimeout:     300, // 5 minutes default
		OperationTimeouts:  make(map[pipeline.OperationType]int),
		RetryPolicy: pipeline.RetryPolicy{
			MaxAttempts: 3,
		},
	}
	
	processor := pipeline.NewSimpleProcessor(db, pipelineConfig, logrus.StandardLogger(), certManager, encryptionKey, eventService)
	
	// Register operation handlers
	userSyncHandler := pipelinehandlers.NewUserSyncHandler(db, logrus.StandardLogger(), certManager, crypto.NewEncryptor(encryptionKey))
	processor.RegisterHandler(pipeline.OpTypeUserSync, userSyncHandler)
	
	// TODO: Register other handlers for each operation type
	// processor.RegisterHandler(pipeline.OpTypeSafeProvision, &handlers.SafeProvisionHandler{})
	// processor.RegisterHandler(pipeline.OpTypeAccessGrant, &handlers.AccessGrantHandler{})
	// etc.
	
	// Start the processor
	if err := processor.Start(ctx); err != nil {
		logrus.WithError(err).Fatal("Failed to start pipeline processor")
	}
	
	cyberarkHandler := handlers.NewCyberArkInstancesHandler(db, logrus.StandardLogger(), encryptionKey, certManager)
	certAuthHandler := handlers.NewCertificateAuthoritiesHandler(db, logrus.StandardLogger(), certManager)
	operationsHandler := handlers.NewOperationsHandler(db, logrus.StandardLogger(), eventService)
	syncSchedulesHandler := handlers.NewSyncSchedulesHandler(db, logrus.StandardLogger(), eventService)
	syncJobsHandler := handlers.NewSyncJobsHandler(db, logrus.StandardLogger(), syncJobService, eventService)
	activityHandler := handlers.NewActivityHandler(db, logrus.StandardLogger(), eventService)

	// API routes
	api := router.Group("/api")
	{
		// Auth routes
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/login/cli", authHandler.LoginCLI) // CLI login endpoint that returns token
		api.POST("/auth/logout", authHandler.Logout)
		
		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthRequired(authHandler))
		{
			// Session check endpoint
			protected.GET("/auth/check", func(c *gin.Context) {
				user := middleware.GetUser(c)
				c.JSON(http.StatusOK, gin.H{
					"authenticated": true,
					"user": user,
				})
			})
			
			// Current user endpoint (used by frontend)
			protected.GET("/auth/me", func(c *gin.Context) {
				user := middleware.GetUser(c)
				c.JSON(http.StatusOK, user)
			})
			
			// Operations routes
			protected.GET("/operations", operationsHandler.ListOperations)
			protected.GET("/operations/:id", operationsHandler.GetOperation)
			protected.POST("/operations", operationsHandler.CreateOperation)
			protected.POST("/operations/:id/cancel", operationsHandler.CancelOperation)
			protected.PATCH("/operations/:id/priority", operationsHandler.UpdatePriority)
			protected.GET("/operations/stream", operationsHandler.StreamOperations)
			
			// CyberArk instances routes
			protected.GET("/cyberark/instances", cyberarkHandler.ListInstances)
			protected.GET("/cyberark/instances/:id", cyberarkHandler.GetInstance)
			protected.POST("/cyberark/instances", cyberarkHandler.CreateInstance)
			protected.PUT("/cyberark/instances/:id", cyberarkHandler.UpdateInstance)
			protected.DELETE("/cyberark/instances/:id", cyberarkHandler.DeleteInstance)
			protected.POST("/cyberark/test-connection", cyberarkHandler.TestConnection)
			protected.POST("/cyberark/instances/:id/test", cyberarkHandler.TestInstanceConnection)
			
			// Certificate Authorities routes
			protected.GET("/certificate-authorities", certAuthHandler.List)
			protected.GET("/certificate-authorities/:id", certAuthHandler.Get)
			protected.POST("/certificate-authorities", certAuthHandler.Create)
			protected.PUT("/certificate-authorities/:id", certAuthHandler.Update)
			protected.DELETE("/certificate-authorities/:id", certAuthHandler.Delete)
			protected.POST("/certificate-authorities/refresh", certAuthHandler.RefreshPool)
			
			// Global sync management routes (deprecated but kept for compatibility)
			protected.GET("/sync/schedules", syncSchedulesHandler.GetSchedules)
			protected.POST("/sync/schedules/pause-all", syncSchedulesHandler.PauseAll)
			protected.POST("/sync/schedules/resume-all", syncSchedulesHandler.ResumeAll)
			
			// Instance-specific sync schedule routes (deprecated but moved to proper location)
			protected.PUT("/instances/:instance_id/sync-schedules", syncSchedulesHandler.UpdateInstanceSchedule)
			protected.PUT("/instances/:instance_id/sync-schedules/:entity_type", syncSchedulesHandler.UpdateInstanceEntitySchedule)
			protected.POST("/instances/:instance_id/sync-schedules/:entity_type/trigger", syncSchedulesHandler.TriggerInstanceSync)
			protected.PUT("/instances/:instance_id/sync-schedules/pause", syncSchedulesHandler.PauseInstance)
			protected.PUT("/instances/:instance_id/sync-schedules/resume", syncSchedulesHandler.ResumeInstance)
			
			// Global sync job routes (for cross-instance views)
			protected.GET("/sync-jobs/:id", syncJobsHandler.GetSyncJob)
			protected.GET("/sync-jobs/stream", syncJobsHandler.StreamSyncJobs)
			
			// Instance-specific sync routes
			protected.GET("/instances/:instance_id/sync-jobs", syncJobsHandler.ListSyncJobsForInstance)
			protected.POST("/instances/:instance_id/sync-jobs/trigger", syncJobsHandler.TriggerSyncForInstance)
			protected.GET("/instances/:instance_id/sync-configs", syncJobsHandler.GetSyncConfigs)
			protected.PATCH("/instances/:instance_id/sync-configs/:sync_type", syncJobsHandler.UpdateSyncConfig)
			
			// Activity routes (unified view)
			protected.GET("/activity", activityHandler.ListActivity)
			protected.GET("/activity/stream", activityHandler.StreamActivity)
			
			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminRequiredGorm())
			{
				// Admin-only endpoints can be added here
			}
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	// Serve static files
	// In development, this returns 404 and frontend is served by Vite
	// In production, embed.go provides the actual implementation
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found - in development mode, use Vite dev server on port 5173"})
	})

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// Start background task to clean up expired sessions
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				if err := authHandler.DeleteExpiredSessions(ctx); err != nil {
					logrus.WithError(err).Error("Failed to delete expired sessions")
				} else {
					logrus.Info("Cleaned up expired sessions")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Graceful shutdown
	go func() {
		logrus.WithField("port", cfg.Server.Port).Info("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down server...")

	// Cancel context to signal shutdown
	cancel()
	
	// Stop the pipeline processor
	logrus.Info("Stopping pipeline processor...")
	if err := processor.Stop(); err != nil {
		logrus.WithError(err).Error("Failed to stop pipeline processor gracefully")
	}
	
	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logrus.WithError(err).Fatal("Server forced to shutdown")
	}

	logrus.Info("Server exited")
}

// handlePasswordReset handles resetting a user's password
func handlePasswordReset(username, newPassword string) error {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable not set")
	}

	// Connect to database
	dbConfig := database.DatabaseConfig{
		Driver: config.GetDatabaseDriver(),
		DSN:    dbURL,
	}

	db, err := database.NewGormConnection(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Update the password
	ctx := context.Background()
	if err := db.UpdateUserPassword(ctx, username, newPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}