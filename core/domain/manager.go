// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package domain provides custom domain management with HTTPS support.
// It handles domain configuration, self-signed certificate generation for
// local development, and integration with Let's Encrypt for production.
package domain

import (
	"fmt"
	"strings"

	"github.com/DoniLite/Mogoly/core/config"
	"github.com/DoniLite/Mogoly/core/events"
	"github.com/DoniLite/Mogoly/core/hosts"
	"github.com/caddyserver/certmagic"
)

var (
	cache *certmagic.Cache
)

func init() {
	cache = certmagic.NewCache(certmagic.CacheOptions{
		GetConfigForCert: func(cert certmagic.Certificate) (*certmagic.Config, error) {
			return certmagic.New(cache, certmagic.Config{
				OnEvent: events.OnCertManagerEvent,
			}), nil
		},
	})
}

// NewManager creates a new domain manager
func NewManager() (*Manager, error) {
	// Create certs directory
	certsDir, err := config.CreateConfigDir("certs")
	if err != nil {
		return nil, fmt.Errorf("failed to create certs directory: %v", err)
	}

	hostsManager, err := hosts.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create hosts manager: %v", err)
	}

	return &Manager{
		domains:      make(map[string]*Config),
		hostsManager: hostsManager,
		certsDir:     certsDir,
	}, nil
}

// IsLocalDomain checks if a domain is a local development domain
func IsLocalDomain(domain string) bool {
	localTLDs := []string{".local", ".test", ".dev", ".localhost"}
	for _, tld := range localTLDs {
		if strings.HasSuffix(domain, tld) {
			return true
		}
	}
	return domain == "localhost"
}
