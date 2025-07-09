package handlers_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/orca-ng/orca/internal/database"
	"github.com/orca-ng/orca/internal/handlers"
	"github.com/orca-ng/orca/internal/models"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
	"github.com/orca-ng/orca/internal/services"
	"github.com/sirupsen/logrus"
)

// TestCyberArkConnectionWithCertificateChain tests the full flow of:
// 1. Creating a CA chain in the certificate store
// 2. Creating a CyberArk instance that uses that CA
// 3. Verifying the connection works with the custom CA
func TestCyberArkConnectionWithCertificateChain(t *testing.T) {
	// Setup database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(
		&gormmodels.CertificateAuthority{},
		&gormmodels.CyberArkInstance{},
		&gormmodels.User{},
		&gormmodels.Session{},
	)
	require.NoError(t, err)
	
	gormDB := &database.GormDB{DB: db}
	logger := logrus.New()
	
	// Initialize services
	certManager := services.NewCertificateManager(gormDB, logger)
	
	// Setup handlers
	caHandler := handlers.NewCertificateAuthoritiesHandler(gormDB, logger, certManager)
	cyberarkHandler := handlers.NewCyberArkInstancesHandler(gormDB, logger, "test-encryption-key-32-bytes-long!", certManager)
	
	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", &models.User{
			ID:       "usr_test123",
			Username: "testuser",
			IsAdmin:  true,
		})
		c.Next()
	})
	
	api := router.Group("/api")
	api.POST("/certificate-authorities", caHandler.Create)
	api.GET("/certificate-authorities", caHandler.List)
	api.POST("/cyberark/test-connection", cyberarkHandler.TestConnection)
	api.POST("/cyberark/instances", cyberarkHandler.CreateInstance)

	// Test Scenario 1: Direct root CA signing
	t.Run("DirectRootCASigning", func(t *testing.T) {
		// Generate a self-signed root CA
		rootKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		rootTemplate := &x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject: pkix.Name{
				Country:      []string{"US"},
				Organization: []string{"Test Corp"},
				CommonName:   "Test Root CA",
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(365 * 24 * time.Hour),
			KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
			IsCA:                  true,
		}
		
		rootCertDER, _ := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
		rootCert, _ := x509.ParseCertificate(rootCertDER)
		rootPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootCertDER}))

		// Store the root CA
		caReq := models.CreateCertificateAuthorityRequest{
			Name:        "Test Root CA - Direct",
			Description: "Root CA for direct signing",
			Certificate: rootPEM,
		}
		
		jsonBody, _ := json.Marshal(caReq)
		req := httptest.NewRequest("POST", "/api/certificate-authorities", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		// Generate a server certificate signed directly by root
		serverKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		serverTemplate := &x509.Certificate{
			SerialNumber: big.NewInt(2),
			Subject: pkix.Name{
				Country:      []string{"US"},
				Organization: []string{"Test Corp"},
				CommonName:   "cyberark.example.com",
			},
			NotBefore:    time.Now(),
			NotAfter:     time.Now().Add(90 * 24 * time.Hour),
			KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			DNSNames:     []string{"cyberark.example.com", "localhost"},
			IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		}
		
		serverCertDER, _ := x509.CreateCertificate(rand.Reader, serverTemplate, rootCert, &serverKey.PublicKey, rootKey)
		
		// Create a test HTTPS server with the server certificate
		testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/PasswordVault/WebServices/auth/Cyberark/CyberArkAuthenticationService.svc/Logon" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"CyberArkLogonResult": "test-token"})
			}
		}))
		
		testServer.TLS = &tls.Config{
			Certificates: []tls.Certificate{
				{
					Certificate: [][]byte{serverCertDER},
					PrivateKey:  serverKey,
				},
			},
		}
		testServer.StartTLS()
		defer testServer.Close()

		// Test connection with the custom CA
		connReq := models.TestCyberArkConnectionRequest{
			BaseURL:  testServer.URL,
			Username: "testuser",
			Password: "testpass",
		}
		
		jsonBody, _ = json.Marshal(connReq)
		req = httptest.NewRequest("POST", "/api/cyberark/test-connection", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var connResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &connResp)
		assert.True(t, connResp["success"].(bool))
	})

	// Test Scenario 2: Intermediate CA chain
	t.Run("IntermediateCAChain", func(t *testing.T) {
		// Generate root CA
		rootKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		rootTemplate := &x509.Certificate{
			SerialNumber: big.NewInt(100),
			Subject: pkix.Name{
				Country:      []string{"US"},
				Organization: []string{"Test Corp"},
				CommonName:   "Test Chain Root CA",
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
			KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
			BasicConstraintsValid: true,
			IsCA:                  true,
		}
		
		rootCertDER, _ := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
		rootCert, _ := x509.ParseCertificate(rootCertDER)
		
		// Generate intermediate CA
		intKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		intTemplate := &x509.Certificate{
			SerialNumber: big.NewInt(101),
			Subject: pkix.Name{
				Country:      []string{"US"},
				Organization: []string{"Test Corp"},
				CommonName:   "Test Intermediate CA",
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(5 * 365 * 24 * time.Hour),
			KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
			BasicConstraintsValid: true,
			IsCA:                  true,
			MaxPathLen:            0,
			MaxPathLenZero:        true,
		}
		
		intCertDER, _ := x509.CreateCertificate(rand.Reader, intTemplate, rootCert, &intKey.PublicKey, rootKey)
		intCert, _ := x509.ParseCertificate(intCertDER)
		
		// Create certificate chain PEM
		chainPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: intCertDER})) +
			string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootCertDER}))
		
		// Store the certificate chain
		caReq := models.CreateCertificateAuthorityRequest{
			Name:        "Test CA Chain",
			Description: "Root and intermediate CA chain",
			Certificate: chainPEM,
		}
		
		jsonBody, _ := json.Marshal(caReq)
		req := httptest.NewRequest("POST", "/api/certificate-authorities", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var caResp models.CertificateAuthority
		json.Unmarshal(w.Body.Bytes(), &caResp)
		assert.Equal(t, 2, caResp.CertificateCount)
		assert.True(t, caResp.IsIntermediate)
		assert.False(t, caResp.IsRootCA)
		
		// Generate server certificate signed by intermediate
		serverKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		serverTemplate := &x509.Certificate{
			SerialNumber: big.NewInt(102),
			Subject: pkix.Name{
				Country:      []string{"US"},
				Organization: []string{"Test Corp"},
				CommonName:   "cyberark-chain.example.com",
			},
			NotBefore:   time.Now(),
			NotAfter:    time.Now().Add(90 * 24 * time.Hour),
			KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			DNSNames:    []string{"cyberark-chain.example.com", "localhost"},
			IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
		}
		
		serverCertDER, _ := x509.CreateCertificate(rand.Reader, serverTemplate, intCert, &serverKey.PublicKey, intKey)
		
		// Create test HTTPS server with full certificate chain
		testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/PasswordVault/WebServices/auth/Cyberark/CyberArkAuthenticationService.svc/Logon" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"CyberArkLogonResult": "test-token-chain"})
			}
		}))
		
		testServer.TLS = &tls.Config{
			Certificates: []tls.Certificate{
				{
					Certificate: [][]byte{serverCertDER, intCertDER}, // Include intermediate in chain
					PrivateKey:  serverKey,
				},
			},
		}
		testServer.StartTLS()
		defer testServer.Close()
		
		// Test connection with certificate chain
		connReq := models.TestCyberArkConnectionRequest{
			BaseURL:  testServer.URL,
			Username: "testuser",
			Password: "testpass",
		}
		
		jsonBody, _ = json.Marshal(connReq)
		req = httptest.NewRequest("POST", "/api/cyberark/test-connection", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var connResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &connResp)
		assert.True(t, connResp["success"].(bool))
		
		// Create a CyberArk instance with the test server
		instanceReq := models.CreateCyberArkInstanceRequest{
			Name:     "Test CyberArk with Chain",
			BaseURL:  testServer.URL,
			Username: "testuser",
			Password: "testpass",
		}
		
		jsonBody, _ = json.Marshal(instanceReq)
		req = httptest.NewRequest("POST", "/api/cyberark/instances", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var instanceResp models.CyberArkInstanceInfo
		json.Unmarshal(w.Body.Bytes(), &instanceResp)
		assert.Equal(t, "Test CyberArk with Chain", instanceResp.Name)
		assert.NotNil(t, instanceResp.LastTestSuccess)
		assert.True(t, *instanceResp.LastTestSuccess)
	})
}

