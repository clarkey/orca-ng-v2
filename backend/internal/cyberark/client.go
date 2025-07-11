package cyberark

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	
	"github.com/orca-ng/orca/internal/services"
)

// HTTPClientFactory is a function that creates an HTTP client
type HTTPClientFactory func() (*http.Client, error)

// Config holds client configuration
type Config struct {
	BaseURL        string
	Username       string
	Password       string
	SkipTLSVerify  bool
	RequestTimeout time.Duration
	CertManager    *services.CertificateManager
}

// Client represents a CyberArk API client
type Client struct {
	baseURL           string
	username          string
	password          string
	httpClient        *http.Client
	httpClientFactory HTTPClientFactory
	token             string
	skipTLSVerify     bool
	certManager       *services.CertificateManager
}

// NewClient creates a new CyberArk client with configuration
func NewClient(cfg Config) (*Client, error) {
	// Ensure baseURL ends without trailing slash
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	
	// Validate URL
	if err := ValidateURL(cfg.BaseURL); err != nil {
		return nil, err
	}
	
	// Default timeout
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 30 * time.Second
	}
	
	// Create HTTP client with custom TLS config
	httpClient := &http.Client{
		Timeout: cfg.RequestTimeout,
	}
	
	// Configure TLS
	if cfg.SkipTLSVerify || cfg.CertManager != nil {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: cfg.SkipTLSVerify,
		}
		
		// Add custom CA if certificate manager is provided
		if cfg.CertManager != nil {
			pool, err := cfg.CertManager.GetCertPool(context.Background())
			if err == nil && pool != nil {
				tlsConfig.RootCAs = pool
			}
		}
		
		httpClient.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}
	
	return &Client{
		baseURL:       cfg.BaseURL,
		username:      cfg.Username,
		password:      cfg.Password,
		httpClient:    httpClient,
		skipTLSVerify: cfg.SkipTLSVerify,
		certManager:   cfg.CertManager,
	}, nil
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

// Authenticate authenticates with CyberArk and returns the session token
func (c *Client) Authenticate() (string, error) {
	return c.AuthenticateWithContext(context.Background())
}

// AuthenticateWithContext authenticates with CyberArk using the provided context
func (c *Client) AuthenticateWithContext(ctx context.Context) (string, error) {
	// Prepare the authentication request
	authURL := fmt.Sprintf("%s/API/auth/Cyberark/Logon", c.baseURL)
	
	payload := map[string]string{
		"username": c.username,
		"password": c.password,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// Get HTTP client
	httpClient := c.getHTTPClient()
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to CyberArk: %w", err)
	}
	defer resp.Body.Close()
	
	// Check the response
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return "", fmt.Errorf("authentication failed: invalid username or password")
		case http.StatusForbidden:
			return "", fmt.Errorf("authentication failed: user is not authorized")
		case http.StatusNotFound:
			return "", fmt.Errorf("invalid CyberArk URL or API endpoint not found")
		default:
			return "", fmt.Errorf("authentication failed with status code: %d", resp.StatusCode)
		}
	}
	
	// Parse the response to get the token
	// CyberArk v10+ returns just a string token, older versions return JSON object
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	
	bodyStr := string(bodyBytes)
	
	// First try to parse as a plain string token (v10+ format)
	// Remove quotes if present
	if strings.HasPrefix(bodyStr, "\"") && strings.HasSuffix(bodyStr, "\"") {
		token := strings.Trim(bodyStr, "\"")
		if token != "" {
			c.token = token
			return token, nil
		}
	}
	
	// Try to parse as JSON object (older format)
	var authResp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &authResp); err != nil {
		// If it's not JSON, treat the whole response as the token
		token := strings.TrimSpace(bodyStr)
		if token != "" {
			c.token = token
			return token, nil
		}
		return "", fmt.Errorf("failed to parse auth response: %w", err)
	}
	
	// Look for token in JSON response
	if tokenRaw, ok := authResp["CyberArkLogonResult"]; ok {
		if token, ok := tokenRaw.(string); ok && token != "" {
			c.token = token
			return token, nil
		}
	}
	
	return "", fmt.Errorf("no token found in auth response")
}

// TestConnection tests the connection to CyberArk by attempting to authenticate
func (c *Client) TestConnection(ctx context.Context) (bool, string, error) {
	startTime := time.Now()
	
	// Try to authenticate
	token, err := c.AuthenticateWithContext(ctx)
	if err != nil {
		return false, err.Error(), err
	}
	
	// Calculate response time
	responseTime := time.Since(startTime).Milliseconds()
	
	// Log off immediately after successful test
	c.token = token
	if err := c.LogoffWithContext(ctx); err != nil {
		// Log warning but don't fail the test
		// This would normally be logged by the logger if we had one
	}
	
	message := fmt.Sprintf("Successfully connected to CyberArk at %s (Response time: %dms)", c.baseURL, responseTime)
	return true, message, nil
}

// Logoff logs off from CyberArk
func (c *Client) Logoff() error {
	return c.LogoffWithContext(context.Background())
}

// LogoffWithContext logs off from CyberArk using the provided context
func (c *Client) LogoffWithContext(ctx context.Context) error {
	if c.token == "" {
		return nil
	}
	
	logoffURL := fmt.Sprintf("%s/API/auth/Logoff", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "POST", logoffURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create logoff request: %w", err)
	}
	
	req.Header.Set("Authorization", c.token)
	
	// Get HTTP client
	httpClient := c.getHTTPClient()
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to logoff: %w", err)
	}
	defer resp.Body.Close()
	
	c.token = ""
	return nil
}

// GetToken returns the current authentication token
func (c *Client) GetToken() string {
	return c.token
}

// SetToken sets the authentication token (useful for session reuse)
func (c *Client) SetToken(token string) {
	c.token = token
}

// IsAuthenticated checks if the client has an authentication token
func (c *Client) IsAuthenticated() bool {
	return c.token != ""
}

// getHTTPClient returns the appropriate HTTP client
func (c *Client) getHTTPClient() *http.Client {
	if c.httpClientFactory != nil {
		client, err := c.httpClientFactory()
		if err == nil {
			return client
		}
	}
	return c.httpClient
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