package cyberark

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPClientFactory is a function that creates an HTTP client
type HTTPClientFactory func() (*http.Client, error)

// Client represents a CyberArk API client
type Client struct {
	baseURL           string
	username          string
	password          string
	httpClient        *http.Client
	httpClientFactory HTTPClientFactory
	token             string
	skipTLSVerify     bool
}

// NewClient creates a new CyberArk client
func NewClient(baseURL, username, password string) *Client {
	// Ensure baseURL ends without trailing slash
	baseURL = strings.TrimRight(baseURL, "/")
	
	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		skipTLSVerify: false,
	}
}

// NewClientWithTLSConfig creates a new CyberArk client with custom TLS configuration
func NewClientWithTLSConfig(baseURL, username, password string, skipTLSVerify bool) *Client {
	// Ensure baseURL ends without trailing slash
	baseURL = strings.TrimRight(baseURL, "/")
	
	// Create HTTP client with custom TLS config if needed
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	if skipTLSVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	
	return &Client{
		baseURL:       baseURL,
		username:      username,
		password:      password,
		httpClient:    httpClient,
		skipTLSVerify: skipTLSVerify,
	}
}

// NewClientWithHTTPClientFactory creates a new CyberArk client with a custom HTTP client factory
func NewClientWithHTTPClientFactory(baseURL, username, password string, factory HTTPClientFactory) *Client {
	// Ensure baseURL ends without trailing slash
	baseURL = strings.TrimRight(baseURL, "/")
	
	return &Client{
		baseURL:           baseURL,
		username:          username,
		password:          password,
		httpClientFactory: factory,
	}
}

// TestConnection tests the connection to CyberArk by attempting to authenticate
func (c *Client) TestConnection(ctx context.Context) (bool, string, error) {
	startTime := time.Now()
	
	// Prepare the authentication request
	// Note: baseURL should already include the full path (e.g., https://server/PasswordVault)
	authURL := fmt.Sprintf("%s/API/auth/Cyberark/Logon", c.baseURL)
	
	payload := map[string]string{
		"username": c.username,
		"password": c.password,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, "", fmt.Errorf("failed to marshal auth payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewReader(jsonData))
	if err != nil {
		return false, "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// Get HTTP client (use factory if available)
	httpClient := c.httpClient
	if c.httpClientFactory != nil {
		client, err := c.httpClientFactory()
		if err != nil {
			return false, "", fmt.Errorf("failed to create HTTP client: %w", err)
		}
		httpClient = client
	}
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to connect to CyberArk: %w", err)
	}
	defer resp.Body.Close()
	
	// Calculate response time
	responseTime := time.Since(startTime).Milliseconds()
	
	// Check the response
	if resp.StatusCode == http.StatusOK {
		// Parse the response to get the token
		var authResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&authResp); err == nil {
			if token, ok := authResp["CyberArkLogonResult"].(string); ok {
				c.token = token
				// Log off immediately after successful test
				c.Logoff(ctx)
			}
		}
		
		message := fmt.Sprintf("Successfully connected to CyberArk at %s (Response time: %dms)", c.baseURL, responseTime)
		return true, message, nil
	}
	
	// Handle different error codes
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return false, "Authentication failed: Invalid username or password", nil
	case http.StatusForbidden:
		return false, "Authentication failed: User is not authorized", nil
	case http.StatusNotFound:
		return false, "Invalid CyberArk URL or API endpoint not found", nil
	default:
		return false, fmt.Sprintf("Connection failed with status code: %d", resp.StatusCode), nil
	}
}

// Logoff logs off from CyberArk
func (c *Client) Logoff(ctx context.Context) error {
	if c.token == "" {
		return nil
	}
	
	logoffURL := fmt.Sprintf("%s/API/auth/Logoff", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "POST", logoffURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create logoff request: %w", err)
	}
	
	req.Header.Set("Authorization", c.token)
	
	// Get HTTP client (use factory if available)
	httpClient := c.httpClient
	if c.httpClientFactory != nil {
		client, err := c.httpClientFactory()
		if err != nil {
			return fmt.Errorf("failed to create HTTP client: %w", err)
		}
		httpClient = client
	}
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to logoff: %w", err)
	}
	defer resp.Body.Close()
	
	c.token = ""
	return nil
}

// ValidateURL validates that the provided URL is a valid CyberArk PVWA URL
func ValidateURL(baseURL string) error {
	// Parse the URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}
	
	// Check if it's HTTPS (recommended for production)
	if u.Scheme != "https" && u.Scheme != "http" {
		return fmt.Errorf("URL must use HTTP or HTTPS protocol")
	}
	
	// Check if host is present
	if u.Host == "" {
		return fmt.Errorf("URL must include a host")
	}
	
	return nil
}