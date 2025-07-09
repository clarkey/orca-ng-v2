package services_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/orca-ng/orca/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to generate a private key
func generatePrivateKey(t *testing.T) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

// Helper function to create a certificate template
func createCertTemplate(commonName string, isCA bool, isSelfSigned bool) *x509.Certificate {
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Country:            []string{"US"},
			Organization:       []string{"Test Organization"},
			OrganizationalUnit: []string{"Test Unit"},
			CommonName:         commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign | x509.KeyUsageCRLSign
		template.ExtKeyUsage = nil // CA certificates typically don't have extended key usage
	}

	return template
}

// Helper function to create and sign a certificate
func createCertificate(t *testing.T, template, parent *x509.Certificate, pubKey *rsa.PublicKey, privKey *rsa.PrivateKey) []byte {
	certDER, err := x509.CreateCertificate(rand.Reader, template, parent, pubKey, privKey)
	require.NoError(t, err)
	return certDER
}

// Helper function to encode certificate to PEM
func encodeCertToPEM(certDER []byte) string {
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}))
}

func TestParseSingleCertificate(t *testing.T) {
	// Generate a self-signed root CA
	rootKey := generatePrivateKey(t)
	rootTemplate := createCertTemplate("Test Root CA", true, true)
	rootCertDER := createCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	rootPEM := encodeCertToPEM(rootCertDER)

	// Test parsing
	certService := services.NewCertificateService()
	parsed, err := certService.ParseCertificate(rootPEM)
	
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.Equal(t, "CN=Test Root CA,OU=Test Unit,O=Test Organization,C=US", parsed.Subject)
	assert.Equal(t, parsed.Subject, parsed.Issuer) // Self-signed
	assert.True(t, parsed.IsCA)
	assert.True(t, parsed.IsSelfSigned)
	assert.NotEmpty(t, parsed.Fingerprint)
}

func TestParseCertificateChain_RootOnly(t *testing.T) {
	// Generate a self-signed root CA
	rootKey := generatePrivateKey(t)
	rootTemplate := createCertTemplate("Test Root CA", true, true)
	rootCertDER := createCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	rootPEM := encodeCertToPEM(rootCertDER)

	// Test parsing chain with only root
	certService := services.NewCertificateService()
	chain, err := certService.ParseCertificateChain(rootPEM)
	
	assert.NoError(t, err)
	assert.NotNil(t, chain)
	assert.Len(t, chain.Certificates, 1)
	assert.Len(t, chain.RootCerts, 1)
	assert.Len(t, chain.IntermediateCerts, 0)
	assert.True(t, chain.IsCompleteChain)
	assert.Equal(t, chain.PrimaryCert.Subject, "CN=Test Root CA,OU=Test Unit,O=Test Organization,C=US")
}

func TestParseCertificateChain_RootAndIntermediate(t *testing.T) {
	// Generate root CA
	rootKey := generatePrivateKey(t)
	rootTemplate := createCertTemplate("Test Root CA", true, true)
	rootCertDER := createCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	rootCert, err := x509.ParseCertificate(rootCertDER)
	require.NoError(t, err)

	// Generate intermediate CA signed by root
	intermediateKey := generatePrivateKey(t)
	intermediateTemplate := createCertTemplate("Test Intermediate CA", true, false)
	intermediateCertDER := createCertificate(t, intermediateTemplate, rootCert, &intermediateKey.PublicKey, rootKey)

	// Create PEM chain (intermediate first, then root)
	chainPEM := encodeCertToPEM(intermediateCertDER) + encodeCertToPEM(rootCertDER)

	// Test parsing
	certService := services.NewCertificateService()
	chain, err := certService.ParseCertificateChain(chainPEM)
	
	assert.NoError(t, err)
	assert.NotNil(t, chain)
	assert.Len(t, chain.Certificates, 2)
	assert.Len(t, chain.RootCerts, 1)
	assert.Len(t, chain.IntermediateCerts, 1)
	assert.True(t, chain.IsCompleteChain)
	
	// Primary cert should be the first one (intermediate)
	assert.Equal(t, chain.PrimaryCert.Subject, "CN=Test Intermediate CA,OU=Test Unit,O=Test Organization,C=US")
	assert.False(t, chain.PrimaryCert.IsSelfSigned)
	assert.True(t, chain.PrimaryCert.IsCA)
}

func TestParseCertificateChain_CompleteChainWithServiceCert(t *testing.T) {
	// Generate root CA
	rootKey := generatePrivateKey(t)
	rootTemplate := createCertTemplate("Test Root CA", true, true)
	rootCertDER := createCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	rootCert, err := x509.ParseCertificate(rootCertDER)
	require.NoError(t, err)

	// Generate intermediate CA signed by root
	intermediateKey := generatePrivateKey(t)
	intermediateTemplate := createCertTemplate("Test Intermediate CA", true, false)
	intermediateCertDER := createCertificate(t, intermediateTemplate, rootCert, &intermediateKey.PublicKey, rootKey)
	intermediateCert, err := x509.ParseCertificate(intermediateCertDER)
	require.NoError(t, err)

	// Generate service certificate signed by intermediate
	serviceKey := generatePrivateKey(t)
	serviceTemplate := createCertTemplate("test.example.com", false, false)
	serviceCertDER := createCertificate(t, serviceTemplate, intermediateCert, &serviceKey.PublicKey, intermediateKey)

	// Create PEM chain (service, intermediate, root)
	chainPEM := encodeCertToPEM(serviceCertDER) + encodeCertToPEM(intermediateCertDER) + encodeCertToPEM(rootCertDER)

	// Test parsing - should fail because service cert is not a CA
	certService := services.NewCertificateService()
	_, err = certService.ParseCertificateChain(chainPEM)
	assert.NoError(t, err) // ParseCertificateChain doesn't validate, just parses
	
	// But validation should fail
	err = certService.ValidateCertificateChain(chainPEM)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not a CA certificate")
}

func TestValidateCertificate_ValidCA(t *testing.T) {
	// Generate a valid CA certificate
	caKey := generatePrivateKey(t)
	caTemplate := createCertTemplate("Test CA", true, true)
	caCertDER := createCertificate(t, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	caPEM := encodeCertToPEM(caCertDER)

	// Test validation
	certService := services.NewCertificateService()
	err := certService.ValidateCertificate(caPEM)
	assert.NoError(t, err)
}

func TestValidateCertificate_NonCA(t *testing.T) {
	// Generate a non-CA certificate
	key := generatePrivateKey(t)
	template := createCertTemplate("test.example.com", false, true)
	certDER := createCertificate(t, template, template, &key.PublicKey, key)
	certPEM := encodeCertToPEM(certDER)

	// Test validation - should fail
	certService := services.NewCertificateService()
	err := certService.ValidateCertificate(certPEM)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a CA certificate")
}

func TestValidateCertificateChain_ValidChain(t *testing.T) {
	// Generate root CA
	rootKey := generatePrivateKey(t)
	rootTemplate := createCertTemplate("Test Root CA", true, true)
	rootCertDER := createCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	rootCert, err := x509.ParseCertificate(rootCertDER)
	require.NoError(t, err)

	// Generate intermediate CA signed by root
	intermediateKey := generatePrivateKey(t)
	intermediateTemplate := createCertTemplate("Test Intermediate CA", true, false)
	intermediateCertDER := createCertificate(t, intermediateTemplate, rootCert, &intermediateKey.PublicKey, rootKey)

	// Create PEM chain
	chainPEM := encodeCertToPEM(intermediateCertDER) + encodeCertToPEM(rootCertDER)

	// Test validation
	certService := services.NewCertificateService()
	err = certService.ValidateCertificateChain(chainPEM)
	assert.NoError(t, err)
}

func TestValidateCertificateChain_IncompleteChain(t *testing.T) {
	// Generate root CA
	rootKey := generatePrivateKey(t)
	rootTemplate := createCertTemplate("Test Root CA", true, true)
	rootCertDER := createCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	rootCert, err := x509.ParseCertificate(rootCertDER)
	require.NoError(t, err)

	// Generate intermediate CA signed by root
	intermediateKey := generatePrivateKey(t)
	intermediateTemplate := createCertTemplate("Test Intermediate CA", true, false)
	intermediateCertDER := createCertificate(t, intermediateTemplate, rootCert, &intermediateKey.PublicKey, rootKey)
	intermediateCert, err := x509.ParseCertificate(intermediateCertDER)
	require.NoError(t, err)

	// Generate second intermediate signed by first intermediate
	intermediate2Key := generatePrivateKey(t)
	intermediate2Template := createCertTemplate("Test Intermediate CA 2", true, false)
	intermediate2CertDER := createCertificate(t, intermediate2Template, intermediateCert, &intermediate2Key.PublicKey, intermediateKey)

	// Create incomplete chain (missing root)
	chainPEM := encodeCertToPEM(intermediate2CertDER) + encodeCertToPEM(intermediateCertDER)

	// Test validation - should fail because root is missing
	certService := services.NewCertificateService()
	err = certService.ValidateCertificateChain(chainPEM)
	assert.Error(t, err)
}

func TestGetCertificatePool_SingleCert(t *testing.T) {
	// Generate a CA certificate
	caKey := generatePrivateKey(t)
	caTemplate := createCertTemplate("Test CA", true, true)
	caCertDER := createCertificate(t, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	caPEM := encodeCertToPEM(caCertDER)

	// Test pool creation
	certService := services.NewCertificateService()
	pool, err := certService.GetCertificatePool([]string{caPEM})
	
	assert.NoError(t, err)
	assert.NotNil(t, pool)
}

func TestGetCertificatePool_WithChain(t *testing.T) {
	// Generate root and intermediate
	rootKey := generatePrivateKey(t)
	rootTemplate := createCertTemplate("Test Root CA", true, true)
	rootCertDER := createCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	rootCert, err := x509.ParseCertificate(rootCertDER)
	require.NoError(t, err)

	intermediateKey := generatePrivateKey(t)
	intermediateTemplate := createCertTemplate("Test Intermediate CA", true, false)
	intermediateCertDER := createCertificate(t, intermediateTemplate, rootCert, &intermediateKey.PublicKey, rootKey)

	// Create chain PEM
	chainPEM := encodeCertToPEM(intermediateCertDER) + encodeCertToPEM(rootCertDER)

	// Test pool creation
	certService := services.NewCertificateService()
	pool, err := certService.GetCertificatePool([]string{chainPEM})
	
	assert.NoError(t, err)
	assert.NotNil(t, pool)
	// Pool should contain both certificates
}

func TestCertificateChainScenarios(t *testing.T) {
	certService := services.NewCertificateService()

	t.Run("DirectlySignedByRoot", func(t *testing.T) {
		// Scenario 1: Service cert directly signed by root CA
		rootKey := generatePrivateKey(t)
		rootTemplate := createCertTemplate("Direct Root CA", true, true)
		rootCertDER := createCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
		rootPEM := encodeCertToPEM(rootCertDER)

		// Store just the root CA
		chain, err := certService.ParseCertificateChain(rootPEM)
		require.NoError(t, err)
		assert.Len(t, chain.Certificates, 1)
		assert.True(t, chain.Certificates[0].IsSelfSigned)
		assert.True(t, chain.IsCompleteChain)
	})

	t.Run("SignedViaIntermediateChain", func(t *testing.T) {
		// Scenario 2: Service cert signed via intermediate chain
		rootKey := generatePrivateKey(t)
		rootTemplate := createCertTemplate("Chain Root CA", true, true)
		rootCertDER := createCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
		rootCert, err := x509.ParseCertificate(rootCertDER)
		require.NoError(t, err)

		// Create intermediate
		intKey := generatePrivateKey(t)
		intTemplate := createCertTemplate("Chain Intermediate CA", true, false)
		intCertDER := createCertificate(t, intTemplate, rootCert, &intKey.PublicKey, rootKey)

		// Store the full chain
		chainPEM := encodeCertToPEM(intCertDER) + encodeCertToPEM(rootCertDER)
		chain, err := certService.ParseCertificateChain(chainPEM)
		require.NoError(t, err)
		assert.Len(t, chain.Certificates, 2)
		assert.Len(t, chain.IntermediateCerts, 1)
		assert.Len(t, chain.RootCerts, 1)
		assert.True(t, chain.IsCompleteChain)
	})

	t.Run("MultipleIntermediates", func(t *testing.T) {
		// Scenario 3: Multiple intermediate CAs
		rootKey := generatePrivateKey(t)
		rootTemplate := createCertTemplate("Multi Root CA", true, true)
		rootCertDER := createCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
		rootCert, err := x509.ParseCertificate(rootCertDER)
		require.NoError(t, err)

		// First intermediate
		int1Key := generatePrivateKey(t)
		int1Template := createCertTemplate("Multi Intermediate CA 1", true, false)
		int1CertDER := createCertificate(t, int1Template, rootCert, &int1Key.PublicKey, rootKey)
		int1Cert, err := x509.ParseCertificate(int1CertDER)
		require.NoError(t, err)

		// Second intermediate
		int2Key := generatePrivateKey(t)
		int2Template := createCertTemplate("Multi Intermediate CA 2", true, false)
		int2CertDER := createCertificate(t, int2Template, int1Cert, &int2Key.PublicKey, int1Key)

		// Store the full chain
		chainPEM := encodeCertToPEM(int2CertDER) + encodeCertToPEM(int1CertDER) + encodeCertToPEM(rootCertDER)
		chain, err := certService.ParseCertificateChain(chainPEM)
		require.NoError(t, err)
		assert.Len(t, chain.Certificates, 3)
		assert.Len(t, chain.IntermediateCerts, 2)
		assert.Len(t, chain.RootCerts, 1)
		assert.True(t, chain.IsCompleteChain)
	})
}

func TestCertificateVerificationWithPool(t *testing.T) {
	certService := services.NewCertificateService()

	// Create a complete chain
	rootKey := generatePrivateKey(t)
	rootTemplate := createCertTemplate("Verify Root CA", true, true)
	rootCertDER := createCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	rootCert, err := x509.ParseCertificate(rootCertDER)
	require.NoError(t, err)

	intKey := generatePrivateKey(t)
	intTemplate := createCertTemplate("Verify Intermediate CA", true, false)
	intCertDER := createCertificate(t, intTemplate, rootCert, &intKey.PublicKey, rootKey)
	intCert, err := x509.ParseCertificate(intCertDER)
	require.NoError(t, err)

	// Create a service certificate
	serviceKey := generatePrivateKey(t)
	serviceTemplate := createCertTemplate("service.example.com", false, false)
	serviceTemplate.DNSNames = []string{"service.example.com", "www.example.com"}
	serviceCertDER := createCertificate(t, serviceTemplate, intCert, &serviceKey.PublicKey, intKey)
	serviceCert, err := x509.ParseCertificate(serviceCertDER)
	require.NoError(t, err)

	// Create certificate pool with our CA chain
	chainPEM := encodeCertToPEM(intCertDER) + encodeCertToPEM(rootCertDER)
	pool, err := certService.GetCertificatePool([]string{chainPEM})
	require.NoError(t, err)

	// Verify the service certificate against our pool
	opts := x509.VerifyOptions{
		DNSName: "service.example.com",
		Roots:   pool,
		Intermediates: pool, // In real usage, intermediates would be separate
	}
	
	chains, err := serviceCert.Verify(opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, chains)
	assert.Len(t, chains[0], 3) // service -> intermediate -> root
}