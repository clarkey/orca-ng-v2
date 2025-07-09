package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	
	"github.com/sirupsen/logrus"
)

// GetInitialAdminCredentials returns the initial admin username and password
// from environment variables or generates secure defaults
func GetInitialAdminCredentials() (username, password string, isGenerated bool) {
	username = os.Getenv("INITIAL_ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}
	
	password = os.Getenv("INITIAL_ADMIN_PASSWORD")
	if password == "" {
		// Generate a secure random password
		password = GenerateSecurePassword(16)
		isGenerated = true
		
		// Log the generated password ONLY on first run
		// In production, this should be logged to a secure location
		logrus.WithFields(logrus.Fields{
			"username": username,
			"generated_password": password,
			"warning": "SAVE THIS PASSWORD - IT WILL NOT BE SHOWN AGAIN",
		}).Warn("Generated initial admin password")
	}
	
	return username, password, isGenerated
}

// GenerateSecurePassword generates a cryptographically secure random password
func GenerateSecurePassword(length int) string {
	// Use a character set that's easy to type but secure
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("failed to generate secure password: %v", err))
	}
	
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	
	return string(bytes)
}

// GenerateSecureKey generates a secure random key for encryption
func GenerateSecureKey() string {
	key := make([]byte, 32) // 256-bit key
	if _, err := rand.Read(key); err != nil {
		panic(fmt.Sprintf("failed to generate secure key: %v", err))
	}
	
	return base64.StdEncoding.EncodeToString(key)
}

// ValidateProductionConfig checks if production configuration is secure
func ValidateProductionConfig() error {
	if os.Getenv("APP_ENV") != "production" {
		return nil // Skip validation for non-production
	}
	
	// Check encryption key
	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" || encKey == "development-key-do-not-use-in-prod-32bytes!!" {
		return fmt.Errorf("ENCRYPTION_KEY must be set to a secure value in production")
	}
	
	// Check session secret
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" || sessionSecret == "development-secret-key-32-bytes!!" {
		return fmt.Errorf("SESSION_SECRET must be set to a secure value in production")
	}
	
	// Validate key lengths
	if len(encKey) < 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be at least 32 characters")
	}
	
	if len(sessionSecret) < 32 {
		return fmt.Errorf("SESSION_SECRET must be at least 32 characters")
	}
	
	return nil
}