package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/orca-ng/orca/internal/config"
	"github.com/orca-ng/orca/internal/database"
	"github.com/orca-ng/orca/internal/handlers"
	"github.com/orca-ng/orca/internal/middleware"
	"github.com/orca-ng/orca/internal/services"
	"github.com/sirupsen/logrus"
)

// Static files embedding is handled in embed.go and embed_dev.go

func main() {
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
	
	cyberarkHandler := handlers.NewCyberArkInstancesHandler(db, logrus.StandardLogger(), encryptionKey, certManager)
	certAuthHandler := handlers.NewCertificateAuthoritiesHandler(db, logrus.StandardLogger(), certManager)
	operationsHandler := handlers.NewOperationsHandler(db, logrus.StandardLogger())

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

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logrus.WithError(err).Fatal("Server forced to shutdown")
	}

	logrus.Info("Server exited")
}