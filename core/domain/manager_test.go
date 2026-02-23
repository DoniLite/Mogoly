// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package domain

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"testing"
)

func TestNewManager(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	err := os.Setenv("HOME", tempDir)
	if err != nil {
		_ = os.Setenv("HOME", originalHome)
		t.Fatalf("Failed to set HOME environment variable: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if m == nil {
		t.Fatal("Manager is nil")
	}

	if m.domains == nil {
		t.Error("Domains map not initialized")
	}

	if m.hostsManager == nil {
		t.Error("Hosts manager not initialized")
	}

	if m.certsDir == "" {
		t.Error("Certs directory not set")
	}

	// Check certs directory was created
	if _, err := os.Stat(m.certsDir); os.IsNotExist(err) {
		t.Errorf("Certs directory not created: %s", m.certsDir)
	}
}

func TestGenerateSelfSignedCert(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	err := os.Setenv("HOME", tempDir)
	if err != nil {
		_ = os.Setenv("HOME", originalHome)
		t.Fatalf("Failed to set HOME environment variable: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	domain := "test.local"
	certPath, keyPath, err := m.generateSelfSignedCert(domain)
	if err != nil {
		t.Fatalf("Failed to generate self-signed certificate: %v", err)
	}

	// Check certificate file exists
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Errorf("Certificate file not created: %s", certPath)
	}

	// Check key file exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("Key file not created: %s", keyPath)
	}

	// Load and verify certificate
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		t.Fatalf("Failed to load generated certificate: %v", err)
	}

	// Parse certificate
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Verify domain name
	if x509Cert.Subject.CommonName != domain {
		t.Errorf("Certificate CN = %s, want %s", x509Cert.Subject.CommonName, domain)
	}

	// Verify DNS names
	found := false
	for _, dnsName := range x509Cert.DNSNames {
		if dnsName == domain {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Domain %s not found in certificate DNS names: %v", domain, x509Cert.DNSNames)
	}
}

func TestAddProductionDomain(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	err := os.Setenv("HOME", tempDir)
	if err != nil {
		_ = os.Setenv("HOME", originalHome)
		t.Fatalf("Failed to set HOME environment variable: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	domain := "api.example.com"
	err = m.Add(domain, false)
	if err != nil {
		t.Fatalf("Failed to add production domain: %v", err)
	}

	config, exists := m.Get(domain)
	if !exists {
		t.Fatal("Domain not found in manager")
	}

	if config.Domain != domain {
		t.Errorf("Domain name = %s, want %s", config.Domain, domain)
	}

	if config.IsLocal {
		t.Error("Domain should not be marked as local")
	}

	if !config.AutoSSL {
		t.Error("AutoSSL should be enabled for production domains")
	}
}

func TestRemoveDomain(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	err := os.Setenv("HOME", tempDir)
	if err != nil {
		_ = os.Setenv("HOME", originalHome)
		t.Fatalf("Failed to set HOME environment variable: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	domain := "api.example.com"

	// Add domain
	err = m.Add(domain, false)
	if err != nil {
		t.Fatalf("Failed to add domain: %v", err)
	}

	// Verify it exists
	_, exists := m.Get(domain)
	if !exists {
		t.Fatal("Domain not found after adding")
	}

	// Remove domain
	err = m.Remove(domain)
	if err != nil {
		t.Fatalf("Failed to remove domain: %v", err)
	}

	// Verify it's gone
	_, exists = m.Get(domain)
	if exists {
		t.Error("Domain still exists after removal")
	}
}

func TestListDomains(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	err := os.Setenv("HOME", tempDir)
	if err != nil {
		_ = os.Setenv("HOME", originalHome)
		t.Fatalf("Failed to set HOME environment variable: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Initially empty
	domains := m.List()
	if len(domains) != 0 {
		t.Errorf("Expected 0 domains, got %d", len(domains))
	}

	// Add production domains
	err = m.Add("api.example.com", false)
	if err != nil {
		t.Fatalf("Failed to add domain: %v", err)
	}
	err = m.Add("staging.example.com", false)
	if err != nil {
		t.Fatalf("Failed to add domain: %v", err)
	}

	domains = m.List()
	if len(domains) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(domains))
	}
}

func TestIsLocalDomain(t *testing.T) {
	tests := []struct {
		domain string
		want   bool
	}{
		{"app.local", true},
		{"api.test", true},
		{"dev.dev", true},
		{"localhost", true},
		{"myapp.localhost", true},
		{"api.example.com", false},
		{"google.com", false},
		{"app.io", false},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got := IsLocalDomain(tt.domain)
			if got != tt.want {
				t.Errorf("IsLocalDomain(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}

func TestConcurrency(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	err := 	os.Setenv("HOME", tempDir)
	if err != nil {
		_ = os.Setenv("HOME", originalHome)
		t.Fatalf("Failed to set HOME environment variable: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	done := make(chan bool)

	// Concurrent domain additions
	for i := 0; i < 10; i++ {
		go func(index int) {
			domain := "api" + string(rune('0'+index)) + ".example.com"
			err := m.Add(domain, false)
			if err != nil {
				t.Errorf("Failed to add entry: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	domains := m.List()
	if len(domains) != 10 {
		t.Errorf("Expected 10 domains after concurrent adds, got %d", len(domains))
	}
}
