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

	// Connect to database
	db, err := database.New(ctx, cfg.Database.URL)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Initialize default admin user if needed
	if err := initializeAdminUser(ctx, db); err != nil {
		logrus.WithError(err).Fatal("Failed to initialize admin user")
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

	// API routes
	api := router.Group("/api")
	{
		// Auth routes
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/login/cli", authHandler.LoginCLI) // CLI login endpoint that returns token
		api.POST("/auth/logout", authHandler.Logout)
		
		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthRequired(db))
		{
			protected.GET("/auth/me", authHandler.GetCurrentUser)
			
			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminRequired())
			{
				// Add admin routes here
			}
		}
	}

	// Serve static files (only in production)
	if os.Getenv("APP_ENV") != "development" {
		setupStaticFiles(router)
	} else {
		// In development, the frontend is served separately
		logrus.Info("Running in development mode, frontend served separately")
	}

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logrus.Info("Shutting down server...")
		
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logrus.WithError(err).Error("Server forced to shutdown")
		}
		
		cancel()
	}()

	logrus.Infof("Starting ORCA server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.WithError(err).Fatal("Failed to start server")
	}

	<-ctx.Done()
	logrus.Info("Server stopped")
}

func isAPIPath(path string) bool {
	return len(path) >= 4 && path[:4] == "/api"
}

func initializeAdminUser(ctx context.Context, db *database.DB) error {
	// Check if admin user exists
	_, err := db.GetUserByUsername(ctx, "admin")
	if err == nil {
		// Admin user already exists
		return nil
	}

	// Create default admin user
	logrus.Info("Creating default admin user")
	defaultPassword := "admin" // Should be changed on first login
	
	_, err = db.CreateUser(ctx, "admin", defaultPassword, true)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	logrus.Warn("Default admin user created with password 'admin' - PLEASE CHANGE THIS PASSWORD")
	return nil
}