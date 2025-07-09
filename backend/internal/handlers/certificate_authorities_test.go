package handlers_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
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

func setupTestDB(t *testing.T) *database.GormDB {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&gormmodels.CertificateAuthority{})
	require.NoError(t, err)

	return &database.GormDB{DB: db}
}

func setupTestRouter(handler *handlers.CertificateAuthoritiesHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Add middleware to inject test user
	router.Use(func(c *gin.Context) {
		c.Set("user", &models.User{
			ID:       "usr_test123",
			Username: "testuser",
			IsAdmin:  true,
		})
		c.Next()
	})

	// Setup routes
	api := router.Group("/api")
	api.GET("/certificate-authorities", handler.List)
	api.GET("/certificate-authorities/:id", handler.Get)
	api.POST("/certificate-authorities", handler.Create)
	api.PUT("/certificate-authorities/:id", handler.Update)
	api.DELETE("/certificate-authorities/:id", handler.Delete)

	return router
}

// Helper functions for certificate generation
func generatePrivateKey(t *testing.T) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func createCertTemplate(commonName string, isCA bool) *x509.Certificate {
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Country:            []string{"US"},
			Organization:       []string{"Test Org"},
			OrganizationalUnit: []string{"Test Unit"},
			CommonName:         commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign | x509.KeyUsageCRLSign
		template.ExtKeyUsage = nil
	}

	return template
}

func createAndSignCertificate(t *testing.T, template, parent *x509.Certificate, pubKey *rsa.PublicKey, privKey *rsa.PrivateKey) []byte {
	certDER, err := x509.CreateCertificate(rand.Reader, template, parent, pubKey, privKey)
	require.NoError(t, err)
	return certDER
}

func encodeCertToPEM(certDER []byte) string {
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}))
}

func TestCreateCertificateAuthority_SingleRootCA(t *testing.T) {
	db := setupTestDB(t)
	logger := logrus.New()
	certManager := services.NewCertificateManager(db, logger)
	handler := handlers.NewCertificateAuthoritiesHandler(db, logger, certManager)
	router := setupTestRouter(handler)

	// Generate a self-signed root CA
	rootKey := generatePrivateKey(t)
	rootTemplate := createCertTemplate("Test Root CA", true)
	rootCertDER := createAndSignCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	rootPEM := encodeCertToPEM(rootCertDER)

	// Create request
	reqBody := models.CreateCertificateAuthorityRequest{
		Name:        "Test Root CA",
		Description: "A test root certificate authority",
		Certificate: rootPEM,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/certificate-authorities", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.CertificateAuthority
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Test Root CA", response.Name)
	assert.Equal(t, 1, response.CertificateCount)
	assert.True(t, response.IsRootCA)
	assert.False(t, response.IsIntermediate)
	assert.NotEmpty(t, response.Fingerprint)
	assert.Contains(t, response.Subject, "CN=Test Root CA")
	assert.Equal(t, response.Subject, response.Issuer) // Self-signed
}

func TestCreateCertificateAuthority_ChainWithIntermediate(t *testing.T) {
	db := setupTestDB(t)
	logger := logrus.New()
	certManager := services.NewCertificateManager(db, logger)
	handler := handlers.NewCertificateAuthoritiesHandler(db, logger, certManager)
	router := setupTestRouter(handler)

	// Generate root CA
	rootKey := generatePrivateKey(t)
	rootTemplate := createCertTemplate("Test Root CA", true)
	rootCertDER := createAndSignCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	rootCert, err := x509.ParseCertificate(rootCertDER)
	require.NoError(t, err)

	// Generate intermediate CA signed by root
	intKey := generatePrivateKey(t)
	intTemplate := createCertTemplate("Test Intermediate CA", true)
	intCertDER := createAndSignCertificate(t, intTemplate, rootCert, &intKey.PublicKey, rootKey)

	// Create chain PEM (intermediate first, then root)
	chainPEM := encodeCertToPEM(intCertDER) + encodeCertToPEM(rootCertDER)

	// Create request
	reqBody := models.CreateCertificateAuthorityRequest{
		Name:        "Test CA Chain",
		Description: "A certificate chain with root and intermediate",
		Certificate: chainPEM,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/certificate-authorities", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.CertificateAuthority
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Test CA Chain", response.Name)
	assert.Equal(t, 2, response.CertificateCount)
	assert.False(t, response.IsRootCA) // Primary cert is intermediate
	assert.True(t, response.IsIntermediate)
	assert.NotEmpty(t, response.ChainInfo)
	assert.Contains(t, response.Subject, "CN=Test Intermediate CA")
	assert.Contains(t, response.Issuer, "CN=Test Root CA")

	// Parse and verify chain info
	var chainInfo []models.CertificateChainInfo
	err = json.Unmarshal([]byte(response.ChainInfo), &chainInfo)
	require.NoError(t, err)
	assert.Len(t, chainInfo, 2)
	
	// First cert should be intermediate
	assert.Contains(t, chainInfo[0].Subject, "CN=Test Intermediate CA")
	assert.False(t, chainInfo[0].IsSelfSigned)
	assert.True(t, chainInfo[0].IsCA)
	
	// Second cert should be root
	assert.Contains(t, chainInfo[1].Subject, "CN=Test Root CA")
	assert.True(t, chainInfo[1].IsSelfSigned)
	assert.True(t, chainInfo[1].IsCA)
}

func TestCreateCertificateAuthority_InvalidCertificate(t *testing.T) {
	db := setupTestDB(t)
	logger := logrus.New()
	certManager := services.NewCertificateManager(db, logger)
	handler := handlers.NewCertificateAuthoritiesHandler(db, logger, certManager)
	router := setupTestRouter(handler)

	// Create request with invalid certificate
	reqBody := models.CreateCertificateAuthorityRequest{
		Name:        "Invalid CA",
		Description: "This should fail",
		Certificate: "-----BEGIN CERTIFICATE-----\nINVALID DATA\n-----END CERTIFICATE-----",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/certificate-authorities", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid certificate")
}

func TestCreateCertificateAuthority_NonCACertificate(t *testing.T) {
	db := setupTestDB(t)
	logger := logrus.New()
	certManager := services.NewCertificateManager(db, logger)
	handler := handlers.NewCertificateAuthoritiesHandler(db, logger, certManager)
	router := setupTestRouter(handler)

	// Generate a non-CA certificate
	key := generatePrivateKey(t)
	template := createCertTemplate("test.example.com", false)
	certDER := createAndSignCertificate(t, template, template, &key.PublicKey, key)
	certPEM := encodeCertToPEM(certDER)

	// Create request
	reqBody := models.CreateCertificateAuthorityRequest{
		Name:        "Non-CA Cert",
		Description: "This should fail",
		Certificate: certPEM,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/certificate-authorities", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "not a CA certificate")
}

func TestListCertificateAuthorities(t *testing.T) {
	db := setupTestDB(t)
	logger := logrus.New()
	certManager := services.NewCertificateManager(db, logger)
	handler := handlers.NewCertificateAuthoritiesHandler(db, logger, certManager)
	router := setupTestRouter(handler)

	// Create test data - a root CA and a chain
	// First, create a root CA
	rootKey := generatePrivateKey(t)
	rootTemplate := createCertTemplate("Root CA 1", true)
	rootCertDER := createAndSignCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	rootPEM := encodeCertToPEM(rootCertDER)

	reqBody1 := models.CreateCertificateAuthorityRequest{
		Name:        "Root CA 1",
		Certificate: rootPEM,
	}
	jsonBody1, _ := json.Marshal(reqBody1)
	req1 := httptest.NewRequest("POST", "/api/certificate-authorities", bytes.NewBuffer(jsonBody1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusCreated, w1.Code)

	// Create a chain with intermediate
	rootKey2 := generatePrivateKey(t)
	rootTemplate2 := createCertTemplate("Root CA 2", true)
	rootCertDER2 := createAndSignCertificate(t, rootTemplate2, rootTemplate2, &rootKey2.PublicKey, rootKey2)
	rootCert2, _ := x509.ParseCertificate(rootCertDER2)

	intKey := generatePrivateKey(t)
	intTemplate := createCertTemplate("Intermediate CA 1", true)
	intCertDER := createAndSignCertificate(t, intTemplate, rootCert2, &intKey.PublicKey, rootKey2)
	chainPEM := encodeCertToPEM(intCertDER) + encodeCertToPEM(rootCertDER2)

	reqBody2 := models.CreateCertificateAuthorityRequest{
		Name:        "CA Chain 1",
		Certificate: chainPEM,
	}
	jsonBody2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest("POST", "/api/certificate-authorities", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusCreated, w2.Code)

	// List all certificate authorities
	req := httptest.NewRequest("GET", "/api/certificate-authorities", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.CertificateAuthority
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response, 2)

	// Verify the certificates
	rootCA := response[0]
	if rootCA.Name != "Root CA 1" {
		rootCA = response[1]
	}
	assert.Equal(t, "Root CA 1", rootCA.Name)
	assert.Equal(t, 1, rootCA.CertificateCount)
	assert.True(t, rootCA.IsRootCA)
	assert.False(t, rootCA.IsIntermediate)

	chainCA := response[0]
	if chainCA.Name != "CA Chain 1" {
		chainCA = response[1]
	}
	assert.Equal(t, "CA Chain 1", chainCA.Name)
	assert.Equal(t, 2, chainCA.CertificateCount)
	assert.False(t, chainCA.IsRootCA)
	assert.True(t, chainCA.IsIntermediate)
}

func TestCertificateChainValidation_ComplexScenarios(t *testing.T) {
	db := setupTestDB(t)
	logger := logrus.New()
	certManager := services.NewCertificateManager(db, logger)
	handler := handlers.NewCertificateAuthoritiesHandler(db, logger, certManager)
	router := setupTestRouter(handler)

	t.Run("ThreeLevelChain", func(t *testing.T) {
		// Create a three-level chain: Root -> Intermediate 1 -> Intermediate 2
		rootKey := generatePrivateKey(t)
		rootTemplate := createCertTemplate("Three Level Root", true)
		rootCertDER := createAndSignCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
		rootCert, _ := x509.ParseCertificate(rootCertDER)

		int1Key := generatePrivateKey(t)
		int1Template := createCertTemplate("Three Level Int 1", true)
		int1CertDER := createAndSignCertificate(t, int1Template, rootCert, &int1Key.PublicKey, rootKey)
		int1Cert, _ := x509.ParseCertificate(int1CertDER)

		int2Key := generatePrivateKey(t)
		int2Template := createCertTemplate("Three Level Int 2", true)
		int2CertDER := createAndSignCertificate(t, int2Template, int1Cert, &int2Key.PublicKey, int1Key)

		// Create chain PEM
		chainPEM := encodeCertToPEM(int2CertDER) + encodeCertToPEM(int1CertDER) + encodeCertToPEM(rootCertDER)

		reqBody := models.CreateCertificateAuthorityRequest{
			Name:        "Three Level Chain",
			Certificate: chainPEM,
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/certificate-authorities", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.CertificateAuthority
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 3, response.CertificateCount)
		assert.Contains(t, response.Subject, "CN=Three Level Int 2")
	})

	t.Run("DuplicateCertificate", func(t *testing.T) {
		// Try to create a certificate authority with the same fingerprint
		rootKey := generatePrivateKey(t)
		rootTemplate := createCertTemplate("Duplicate Test Root", true)
		rootCertDER := createAndSignCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
		rootPEM := encodeCertToPEM(rootCertDER)

		// First creation should succeed
		reqBody := models.CreateCertificateAuthorityRequest{
			Name:        "First Instance",
			Certificate: rootPEM,
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/certificate-authorities", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Second creation with same cert should fail
		reqBody.Name = "Second Instance"
		jsonBody, _ = json.Marshal(reqBody)
		req = httptest.NewRequest("POST", "/api/certificate-authorities", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "already registered")
	})
}