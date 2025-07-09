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
	Certificate  *x509.Certificate
	PEMData      string
	Fingerprint  string
	Subject      string
	Issuer       string
	NotBefore    time.Time
	NotAfter     time.Time
	IsCA         bool
	IsSelfSigned bool
}

type ParsedCertificateChain struct {
	Certificates     []*ParsedCertificate
	PrimaryCert      *ParsedCertificate // The first certificate in the chain
	RootCerts        []*ParsedCertificate // Self-signed root certificates
	IntermediateCerts []*ParsedCertificate // Intermediate CA certificates
	IsCompleteChain  bool                 // True if chain leads to a self-signed root
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
	isSelfSigned := cert.Subject.String() == cert.Issuer.String()
	
	return &ParsedCertificate{
		Certificate:  cert,
		PEMData:      pemData,
		Fingerprint:  fingerprint,
		Subject:      cert.Subject.String(),
		Issuer:       cert.Issuer.String(),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		IsCA:         cert.IsCA,
		IsSelfSigned: isSelfSigned,
	}, nil
}

// ParseCertificateChain parses a PEM-encoded string containing one or more certificates
func (s *CertificateService) ParseCertificateChain(pemData string) (*ParsedCertificateChain, error) {
	pemData = strings.TrimSpace(pemData)
	if pemData == "" {
		return nil, fmt.Errorf("empty certificate data")
	}
	
	result := &ParsedCertificateChain{
		Certificates:      []*ParsedCertificate{},
		RootCerts:         []*ParsedCertificate{},
		IntermediateCerts: []*ParsedCertificate{},
	}
	
	// Parse all certificates in the PEM data
	remaining := []byte(pemData)
	var pemBlocks []string
	
	for len(remaining) > 0 {
		block, rest := pem.Decode(remaining)
		if block == nil {
			break
		}
		remaining = rest
		
		if block.Type != "CERTIFICATE" {
			continue
		}
		
		// Re-encode this single certificate to PEM for storage
		singlePEM := pem.EncodeToMemory(block)
		pemBlocks = append(pemBlocks, string(singlePEM))
		
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate at position %d: %w", len(result.Certificates)+1, err)
		}
		
		fingerprint := s.calculateFingerprint(cert.Raw)
		isSelfSigned := cert.Subject.String() == cert.Issuer.String()
		
		parsed := &ParsedCertificate{
			Certificate:  cert,
			PEMData:      string(singlePEM),
			Fingerprint:  fingerprint,
			Subject:      cert.Subject.String(),
			Issuer:       cert.Issuer.String(),
			NotBefore:    cert.NotBefore,
			NotAfter:     cert.NotAfter,
			IsCA:         cert.IsCA,
			IsSelfSigned: isSelfSigned,
		}
		
		result.Certificates = append(result.Certificates, parsed)
		
		// Categorize the certificate
		if isSelfSigned && cert.IsCA {
			result.RootCerts = append(result.RootCerts, parsed)
		} else if cert.IsCA {
			result.IntermediateCerts = append(result.IntermediateCerts, parsed)
		}
	}
	
	if len(result.Certificates) == 0 {
		return nil, fmt.Errorf("no valid certificates found in PEM data")
	}
	
	// The primary certificate is the first one
	result.PrimaryCert = result.Certificates[0]
	
	// Check if we have a complete chain (ends with a self-signed root)
	result.IsCompleteChain = len(result.RootCerts) > 0
	
	return result, nil
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

// ValidateCertificateChain validates an entire certificate chain
func (s *CertificateService) ValidateCertificateChain(pemData string) error {
	chain, err := s.ParseCertificateChain(pemData)
	if err != nil {
		return err
	}
	
	// Ensure we have at least one certificate
	if len(chain.Certificates) == 0 {
		return fmt.Errorf("no certificates found in chain")
	}
	
	// Validate that all certificates in the chain are CA certificates
	for i, cert := range chain.Certificates {
		if !cert.IsCA {
			return fmt.Errorf("certificate at position %d is not a CA certificate", i+1)
		}
		
		// Check basic constraints
		if cert.Certificate.BasicConstraintsValid && !cert.Certificate.IsCA {
			return fmt.Errorf("certificate at position %d has invalid basic constraints for a CA", i+1)
		}
		
		// Check key usage
		if cert.Certificate.KeyUsage != 0 {
			if cert.Certificate.KeyUsage&x509.KeyUsageCertSign == 0 {
				return fmt.Errorf("certificate at position %d does not have certificate signing permission", i+1)
			}
		}
	}
	
	// Build certificate pools for validation
	roots := x509.NewCertPool()
	intermediates := x509.NewCertPool()
	
	// Add root certificates to root pool
	for _, cert := range chain.RootCerts {
		roots.AddCert(cert.Certificate)
	}
	
	// Add intermediate certificates to intermediate pool
	for _, cert := range chain.IntermediateCerts {
		intermediates.AddCert(cert.Certificate)
	}
	
	// If we have intermediate certificates, verify they chain properly
	if len(chain.IntermediateCerts) > 0 {
		for _, intermediate := range chain.IntermediateCerts {
			opts := x509.VerifyOptions{
				Roots:         roots,
				Intermediates: intermediates,
				KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
			}
			
			_, err := intermediate.Certificate.Verify(opts)
			if err != nil {
				return fmt.Errorf("intermediate certificate '%s' failed validation: %w", 
					intermediate.Subject, err)
			}
		}
	}
	
	return nil
}

func (s *CertificateService) GetCertificatePool(pemCertificates []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	
	for _, pemData := range pemCertificates {
		// Try to parse as a certificate chain first
		chain, err := s.ParseCertificateChain(pemData)
		if err == nil && len(chain.Certificates) > 0 {
			// Add all certificates from the chain
			for _, cert := range chain.Certificates {
				pool.AddCert(cert.Certificate)
			}
		} else {
			// Fall back to parsing as a single certificate
			parsed, err := s.ParseCertificate(pemData)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate: %w", err)
			}
			pool.AddCert(parsed.Certificate)
		}
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
		// Try to parse as a certificate chain first
		chain, err := s.ParseCertificateChain(pemData)
		if err == nil && len(chain.Certificates) > 0 {
			// Add all certificates from the chain
			for _, cert := range chain.Certificates {
				pool.AddCert(cert.Certificate)
			}
		} else {
			// Fall back to parsing as a single certificate
			parsed, err := s.ParseCertificate(pemData)
			if err != nil {
				return nil, fmt.Errorf("failed to parse custom certificate: %w", err)
			}
			pool.AddCert(parsed.Certificate)
		}
	}
	
	return pool, nil
}

func (s *CertificateService) calculateFingerprint(certDER []byte) string {
	hash := sha256.Sum256(certDER)
	return hex.EncodeToString(hash[:])
}