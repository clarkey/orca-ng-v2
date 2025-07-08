package services

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"sync"
	"time"
	
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type CertificateManager struct {
	db              *pgxpool.Pool
	logger          *logrus.Logger
	certService     *CertificateService
	certPool        *x509.CertPool
	certPoolMutex   sync.RWMutex
	lastRefresh     time.Time
	refreshInterval time.Duration
}

func NewCertificateManager(db *pgxpool.Pool, logger *logrus.Logger) *CertificateManager {
	return &CertificateManager{
		db:              db,
		logger:          logger,
		certService:     NewCertificateService(),
		refreshInterval: 5 * time.Minute, // Refresh certificates every 5 minutes
	}
}

func (cm *CertificateManager) GetCertPool(ctx context.Context) (*x509.CertPool, error) {
	cm.certPoolMutex.RLock()
	needsRefresh := time.Since(cm.lastRefresh) > cm.refreshInterval
	cm.certPoolMutex.RUnlock()
	
	if needsRefresh || cm.certPool == nil {
		if err := cm.refreshCertificates(ctx); err != nil {
			cm.logger.WithError(err).Error("Failed to refresh certificates")
			// If we have a cached pool, return it even if refresh failed
			if cm.certPool != nil {
				return cm.certPool, nil
			}
			return nil, err
		}
	}
	
	cm.certPoolMutex.RLock()
	defer cm.certPoolMutex.RUnlock()
	return cm.certPool, nil
}

func (cm *CertificateManager) refreshCertificates(ctx context.Context) error {
	query := `
		SELECT certificate
		FROM certificate_authorities
		WHERE is_active = true AND not_after > CURRENT_TIMESTAMP
	`
	
	rows, err := cm.db.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query active certificates: %w", err)
	}
	defer rows.Close()
	
	var pemCertificates []string
	for rows.Next() {
		var cert string
		if err := rows.Scan(&cert); err != nil {
			cm.logger.WithError(err).Error("Failed to scan certificate")
			continue
		}
		pemCertificates = append(pemCertificates, cert)
	}
	
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}
	
	pool, err := cm.certService.GetSystemAndCustomCertPool(pemCertificates)
	if err != nil {
		return fmt.Errorf("failed to create certificate pool: %w", err)
	}
	
	cm.certPoolMutex.Lock()
	cm.certPool = pool
	cm.lastRefresh = time.Now()
	cm.certPoolMutex.Unlock()
	
	cm.logger.WithField("certificate_count", len(pemCertificates)).Info("Refreshed certificate pool")
	return nil
}

func (cm *CertificateManager) GetTLSConfig(ctx context.Context, skipTLSVerify bool) (*tls.Config, error) {
	if skipTLSVerify {
		return &tls.Config{
			InsecureSkipVerify: true,
		}, nil
	}
	
	certPool, err := cm.GetCertPool(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate pool: %w", err)
	}
	
	return &tls.Config{
		RootCAs: certPool,
	}, nil
}

func (cm *CertificateManager) GetHTTPClient(ctx context.Context, skipTLSVerify bool, timeout time.Duration) (*http.Client, error) {
	tlsConfig, err := cm.GetTLSConfig(ctx, skipTLSVerify)
	if err != nil {
		return nil, err
	}
	
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}, nil
}