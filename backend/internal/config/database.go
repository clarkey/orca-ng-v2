package config

import "os"

// GetDatabaseDriver returns the database driver from environment variable
// Defaults to "postgres" if not set
func GetDatabaseDriver() string {
	driver := os.Getenv("DATABASE_DRIVER")
	if driver == "" {
		return "postgres"
	}
	
	// Validate supported drivers
	switch driver {
	case "postgres", "mysql", "sqlite", "sqlserver":
		return driver
	default:
		// Default to postgres for invalid driver
		return "postgres"
	}
}