package domain

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DoniLite/Mogoly/core/config"
	"github.com/DoniLite/Mogoly/core/events"
	"github.com/caddyserver/certmagic"
)

func getEndPointFromEnvConfig(envKey string) string {
	env := config.GetEnv(envKey, "production")

	if env == "production" {
		return certmagic.LetsEncryptProductionCA
	}

	return certmagic.LetsEncryptStagingCA

}


func (m *Manager) InitCertManager(cacheDir, email, envKey string) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		events.Logf(events.LOG_ERROR, "[CERT MANAGER]: cert cache mkdir error: %s", err.Error())
	}
	cfg := certmagic.New(cache, certmagic.Config{})
	userACME := certmagic.NewACMEIssuer(cfg, certmagic.ACMEIssuer{
		CA:     getEndPointFromEnvConfig(config.MOGOLY_ENV),
		Email:  email,
		Agreed: true,
	})
	storage := &certmagic.FileStorage{Path: cacheDir}
	cfg.Storage = storage
	cfg.Issuers = []certmagic.Issuer{userACME}
	m.cm = cfg
}

// Add adds a custom domain to the system
func (m *Manager) Add(domain string, isLocal bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if domain already exists
	if _, exists := m.domains[domain]; exists {
		return fmt.Errorf("domain %s already exists", domain)
	}

	config := &Config{
		Domain:  domain,
		IsLocal: isLocal,
	}

	if isLocal {
		// Generate self-signed certificate
		certPath, keyPath, err := m.generateSelfSignedCert(domain)
		if err != nil {
			return fmt.Errorf("failed to generate certificate: %v", err)
		}

		config.CertPath = certPath
		config.KeyPath = keyPath

		// Add to hosts file
		if err := m.hostsManager.Add(domain, "127.0.0.1"); err != nil {
			// Rollback: remove certificate files
			os.Remove(certPath)
			os.Remove(keyPath)
			return fmt.Errorf("failed to add hosts entry: %v", err)
		}
	} else {
		// Production domain - will use Let's Encrypt via CertMagic
		config.AutoSSL = true
	}

	m.domains[domain] = config
	return nil
}

// Remove removes a custom domain
func (m *Manager) Remove(domain string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	config, exists := m.domains[domain]
	if !exists {
		return fmt.Errorf("domain %s not found", domain)
	}

	if config.IsLocal {
		// Remove from hosts file
		if err := m.hostsManager.Remove(domain); err != nil {
			return fmt.Errorf("failed to remove hosts entry: %v", err)
		}

		// Remove certificate files
		if config.CertPath != "" {
			os.Remove(config.CertPath)
		}
		if config.KeyPath != "" {
			os.Remove(config.KeyPath)
		}
	}

	delete(m.domains, domain)
	return nil
}

// Get returns domain configuration
func (m *Manager) Get(domain string) (*Config, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	config, exists := m.domains[domain]
	return config, exists
}

// List returns all configured domains
func (m *Manager) List() []*Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	domains := make([]*Config, 0, len(m.domains))
	for _, config := range m.domains {
		domains = append(domains, config)
	}
	return domains
}

// LoadCertificate loads the certificate for a domain
func (m *Manager) LoadCertificate(domain string) (*tls.Certificate, error) {
	m.mu.RLock()
	config, exists := m.domains[domain]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("domain %s not found", domain)
	}

	if config.CertPath == "" || config.KeyPath == "" {
		return nil, fmt.Errorf("no certificate paths for domain %s", domain)
	}

	cert, err := tls.LoadX509KeyPair(config.CertPath, config.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %v", err)
	}

	return &cert, nil
}

func (m *Manager) GetCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if chi == nil || chi.ServerName == "" {
		return nil, errors.New("missing SNI")
	}
	name := strings.ToLower(chi.ServerName)
	if IsLocalDomain(name) {
		return m.LoadCertificate(name)
	}
	// Public domain: Let certmagic manage/renew
	if err := m.cm.ManageSync(context.Background(), []string{name}); err != nil {
		return nil, fmt.Errorf("certmagic ManageSync: %w", err)
	}
	return m.cm.GetCertificate(chi)
}

// GetCertsDir returns the certificates directory path
func (m *Manager) GetCertsDir() string {
	return m.certsDir
}

// generateSelfSignedCert creates a self-signed certificate for a domain
func (m *Manager) generateSelfSignedCert(domain string) (certPath, keyPath string, err error) {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Mogoly Local Development"},
			CommonName:   domain,
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate: %v", err)
	}

	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key to PEM
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Save to files
	certPath = filepath.Join(m.certsDir, domain+".crt")
	keyPath = filepath.Join(m.certsDir, domain+".key")

	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return "", "", fmt.Errorf("failed to write certificate: %v", err)
	}

	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		os.Remove(certPath) // Cleanup cert file
		return "", "", fmt.Errorf("failed to write private key: %v", err)
	}

	return certPath, keyPath, nil
}
