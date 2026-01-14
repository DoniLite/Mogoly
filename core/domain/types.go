package domain

import (
	"sync"

	"github.com/DoniLite/Mogoly/core/hosts"
	"github.com/caddyserver/certmagic"
)

// Config represents a custom domain configuration
type Config struct {
	Domain   string `json:"domain" yaml:"domain"`
	IsLocal  bool   `json:"is_local" yaml:"is_local"`
	CertPath string `json:"cert_path,omitempty" yaml:"cert_path,omitempty"`
	KeyPath  string `json:"key_path,omitempty" yaml:"key_path,omitempty"`
	AutoSSL  bool   `json:"auto_ssl,omitempty" yaml:"auto_ssl,omitempty"`
}

// Manager manages custom domains and certificates
type Manager struct {
	domains      map[string]*Config
	hostsManager *hosts.Manager
	certsDir     string
	mu           sync.RWMutex
	cm           *certmagic.Config
}
