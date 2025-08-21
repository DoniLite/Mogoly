// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"crypto/tls"
	"log"
	"net/http/httputil"
	"net/url"
	"os"

	certmagic "github.com/caddyserver/certmagic"
)

func NewProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	return proxy
}

func NewCertManager(cacheDir, email string) *CertManager {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		log.Printf("cert cache mkdir: %v", err)
	}
	storage := &certmagic.FileStorage{Path: cacheDir}
	cfg := certmagic.NewDefault()
	cfg.Storage = storage
	return &CertManager{cm: cfg, selfStore: make(map[string]*tls.Certificate)}
}
