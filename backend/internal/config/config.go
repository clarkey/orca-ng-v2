package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Session  SessionConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port int
	Host string
}

type DatabaseConfig struct {
	URL string
}

type SessionConfig struct {
	Secret         string
	SessionTimeout int // in minutes
}

type LogConfig struct {
	Level string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetEnvPrefix("ORCA")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("session.timeout", 1440) // 24 hours
	viper.SetDefault("log.level", "info")

	// Override with environment variables
	viper.BindEnv("database.url", "DATABASE_URL")
	viper.BindEnv("session.secret", "SESSION_SECRET")
	viper.BindEnv("log.level", "LOG_LEVEL")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required config
	if config.Database.URL == "" {
		config.Database.URL = os.Getenv("DATABASE_URL")
		if config.Database.URL == "" {
			return nil, fmt.Errorf("database URL is required")
		}
	}

	if config.Session.Secret == "" {
		config.Session.Secret = os.Getenv("SESSION_SECRET")
		if config.Session.Secret == "" {
			return nil, fmt.Errorf("session secret is required")
		}
	}

	return &config, nil
}