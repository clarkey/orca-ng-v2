package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type SessionInfo struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	ServerURL string    `json:"server_url"`
	Username  string    `json:"username"`
}

type SessionStore struct {
	configDir string
}

func NewSessionStore() (*SessionStore, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &SessionStore{
		configDir: configDir,
	}, nil
}

func (s *SessionStore) Save(session *SessionInfo) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	sessionFile := filepath.Join(s.configDir, "session.json")
	if err := os.WriteFile(sessionFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

func (s *SessionStore) Load() (*SessionInfo, error) {
	sessionFile := filepath.Join(s.configDir, "session.json")
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session SessionInfo
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Remove expired session
		s.Delete()
		return nil, nil
	}

	return &session, nil
}

func (s *SessionStore) Delete() error {
	sessionFile := filepath.Join(s.configDir, "session.json")
	if err := os.Remove(sessionFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session file: %w", err)
	}
	return nil
}

func getConfigDir() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			baseDir = os.Getenv("USERPROFILE")
			if baseDir == "" {
				return "", fmt.Errorf("cannot determine config directory on Windows")
			}
			baseDir = filepath.Join(baseDir, "AppData", "Roaming")
		}
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, "Library", "Application Support")
	default: // Linux and other Unix-like systems
		baseDir = os.Getenv("XDG_CONFIG_HOME")
		if baseDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("cannot determine home directory: %w", err)
			}
			baseDir = filepath.Join(homeDir, ".config")
		}
	}

	return filepath.Join(baseDir, "orca-cli"), nil
}