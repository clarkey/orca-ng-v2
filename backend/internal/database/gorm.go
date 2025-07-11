package database

import (
	"fmt"
	"log"
	"os"
	"time"
	
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	
	"github.com/orca-ng/orca/internal/config"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
	"github.com/orca-ng/orca/pkg/crypto"
)

type GormDB struct {
	*gorm.DB
}

type DatabaseConfig struct {
	Driver   string // "postgres", "mysql", "sqlite", "sqlserver"
	DSN      string
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// NewGormConnection creates a new GORM database connection
func NewGormConnection(cfg DatabaseConfig) (*GormDB, error) {
	var dialector gorm.Dialector
	
	switch cfg.Driver {
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	case "mysql":
		dialector = mysql.Open(cfg.DSN)
	case "sqlite":
		dialector = sqlite.Open(cfg.DSN)
	case "sqlserver":
		dialector = sqlserver.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
	
	// Configure GORM logger
	logLevel := logger.Silent
	if os.Getenv("GORM_DEBUG") == "true" {
		logLevel = logger.Info
	}
	
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // Use singular table names
		},
		DisableForeignKeyConstraintWhenMigrating: true, // For cross-database compatibility
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL database: %w", err)
	}
	
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	
	return &GormDB{DB: db}, nil
}

// AutoMigrate runs GORM auto-migration for all models
func (db *GormDB) AutoMigrate() error {
	return db.DB.AutoMigrate(
		&gormmodels.User{},
		&gormmodels.Session{},
		&gormmodels.CertificateAuthority{},
		&gormmodels.CyberArkInstance{},
		&gormmodels.CyberArkUser{},
		&gormmodels.CyberArkGroupMembership{},
		&gormmodels.CyberArkVaultAuthorization{},
		&gormmodels.Operation{},
		&gormmodels.PipelineConfig{},
	)
}

// SeedDefaultData creates default data (like admin user)
func (db *GormDB) SeedDefaultData() error {
	// Check if admin user exists
	var count int64
	if err := db.Model(&gormmodels.User{}).Where("username = ?", "admin").Count(&count).Error; err != nil {
		return err
	}
	
	if count == 0 {
		// Get initial admin credentials from config
		username, password, isGenerated := config.GetInitialAdminCredentials()
		
		hashedPassword, err := crypto.HashPassword(password)
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}
		
		adminUser := &gormmodels.User{
			Username:     username,
			PasswordHash: hashedPassword,
			IsActive:     true,
			IsAdmin:      true,
		}
		
		if err := db.Create(adminUser).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		
		if isGenerated {
			log.Printf("IMPORTANT: Initial admin user created with username: %s", username)
			log.Println("IMPORTANT: Check logs above for the generated password. It will not be shown again!")
		} else {
			log.Printf("Initial admin user created with username: %s", username)
		}
	}
	
	// Initialize pipeline config defaults
	defaultConfigs := []struct {
		Key   string
		Value string
		Desc  string
	}{
		{
			Key:   "worker_count",
			Value: `{"value": 5}`,
			Desc:  "Number of concurrent pipeline workers",
		},
		{
			Key:   "max_retries",
			Value: `{"value": 3}`,
			Desc:  "Maximum retries for failed operations",
		},
		{
			Key:   "poll_interval",
			Value: `{"value": 5}`,
			Desc:  "Pipeline poll interval in seconds",
		},
	}
	
	for _, cfg := range defaultConfigs {
		var count int64
		if err := db.Model(&gormmodels.PipelineConfig{}).Where("key = ?", cfg.Key).Count(&count).Error; err != nil {
			return err
		}
		
		if count == 0 {
			desc := cfg.Desc
			config := &gormmodels.PipelineConfig{
				Key:         cfg.Key,
				Value:       []byte(cfg.Value),
				Description: &desc,
			}
			
			if err := db.Create(config).Error; err != nil {
				return fmt.Errorf("failed to create pipeline config %s: %w", cfg.Key, err)
			}
		}
	}
	
	return nil
}

// WithUserContext adds user ID to the context for audit fields
func (db *GormDB) WithUserContext(userID string) *gorm.DB {
	return db.DB.WithContext(db.DB.Statement.Context)
}

// Close closes the database connection
func (db *GormDB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}