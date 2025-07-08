package services

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"
	"time"
)

type CertificateService struct{}

func NewCertificateService() *CertificateService {
	return &CertificateService{}
}

type ParsedCertificate struct {
	Certificate *x509.Certificate
	PEMData     string
	Fingerprint string
	Subject     string
	Issuer      string
	NotBefore   time.Time
	NotAfter    time.Time
}

func (s *CertificateService) ParseCertificate(pemData string) (*ParsedCertificate, error) {
	pemData = strings.TrimSpace(pemData)
	
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}
	
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("PEM block is not a certificate (found: %s)", block.Type)
	}
	
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}
	
	fingerprint := s.calculateFingerprint(cert.Raw)
	
	return &ParsedCertificate{
		Certificate: cert,
		PEMData:     pemData,
		Fingerprint: fingerprint,
		Subject:     cert.Subject.String(),
		Issuer:      cert.Issuer.String(),
		NotBefore:   cert.NotBefore,
		NotAfter:    cert.NotAfter,
	}, nil
}

func (s *CertificateService) ValidateCertificate(pemData string) error {
	parsed, err := s.ParseCertificate(pemData)
	if err != nil {
		return err
	}
	
	// Check if certificate is a CA certificate
	if !parsed.Certificate.IsCA {
		return fmt.Errorf("certificate is not a CA certificate")
	}
	
	// Check basic constraints
	if parsed.Certificate.BasicConstraintsValid && !parsed.Certificate.IsCA {
		return fmt.Errorf("certificate has invalid basic constraints for a CA")
	}
	
	// Check key usage
	if parsed.Certificate.KeyUsage != 0 {
		if parsed.Certificate.KeyUsage&x509.KeyUsageCertSign == 0 {
			return fmt.Errorf("certificate does not have certificate signing permission")
		}
	}
	
	return nil
}

func (s *CertificateService) GetCertificatePool(pemCertificates []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	
	for _, pemData := range pemCertificates {
		parsed, err := s.ParseCertificate(pemData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}
		
		pool.AddCert(parsed.Certificate)
	}
	
	return pool, nil
}

func (s *CertificateService) GetSystemAndCustomCertPool(customPEMCertificates []string) (*x509.CertPool, error) {
	// Start with system root CAs
	pool, err := x509.SystemCertPool()
	if err != nil {
		// If system cert pool is not available, create a new empty pool
		pool = x509.NewCertPool()
	}
	
	// Add custom certificates
	for _, pemData := range customPEMCertificates {
		parsed, err := s.ParseCertificate(pemData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse custom certificate: %w", err)
		}
		
		pool.AddCert(parsed.Certificate)
	}
	
	return pool, nil
}

func (s *CertificateService) calculateFingerprint(certDER []byte) string {
	hash := sha256.Sum256(certDER)
	return hex.EncodeToString(hash[:])
}