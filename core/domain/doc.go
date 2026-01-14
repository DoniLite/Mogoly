/*
Package domain provides custom domain management with HTTPS support.

# Overview

The domain package enables you to:
  - Create local development domains (.local, .test, .dev) with self-signed certificates
  - Use production domains with automatic Let's Encrypt certificates
  - Bind domains to any local server or service
  - Access local services via HTTPS

# Quick Start

	import "github.com/DoniLite/Mogoly/core/domain"

	func main() {
		// Create domain manager
		manager, err := domain.NewManager()
		if err != nil {
			log.Fatal(err)
		}

		// Add a local domain (requires sudo for /etc/hosts modification)
		err = manager.Add("myapp.local", true)
		if err != nil {
			log.Fatal(err)
		}

		// Load the certificate for your HTTPS server
		cert, err := manager.LoadCertificate("myapp.local")
		if err != nil {
			log.Fatal(err)
		}

		// Use with your HTTP server
		server := &http.Server{
			Addr: ":443",
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{*cert},
			},
		}
		server.ListenAndServeTLS("", "")
	}

# Local Development

For local domains (.local, .test, .dev), the manager:

1. Generates a self-signed certificate (valid for 1 year)
2. Saves cert/key to ~/.mogoly/certs/
3. Adds entry to /etc/hosts: 127.0.0.1 domain

	// Add local domain
	manager.Add("api.local", true)

	// Certificate files created at:
	// ~/.mogoly/certs/api.local.crt
	// ~/.mogoly/certs/api.local.key

	// /etc/hosts entry added:
	// 127.0.0.1 api.local # Managed by Mogoly

# Production

For production domains, use with CertMagic:

	// Add production domain (no hosts file modification)
	manager.Add("api.myapp.com", false)

	// config.AutoSSL will be true
	// Use CertMagic for automatic Let's Encrypt certificates

# Usage Patterns

## Pattern 1: Direct Access

Your local server runs on any port:

	manager.Add("api.local", true)
	// Access: https://api.local:8080

## Pattern 2: With Load Balancer

Use Mogoly's load balancer to proxy:

	manager.Add("api.local", true)
	// Load balancer listens on :443
	// Access: https://api.local → proxied to backends

## Pattern 3: Multiple Domains

	manager.Add("frontend.local", true)  // React app
	manager.Add("backend.local", true)   // API server
	manager.Add("db.local", true)        // Database

# Permission Requirements

	// Adding local domains requires sudo (to modify /etc/hosts):
	// sudo mogoly domain add myapp.local --local

	// Production domains don't need sudo:
	// mogoly domain add api.myapp.com

# See Also

  - hosts.Manager for direct hosts file management
  - core.CertManager for certificate management
*/
package domain
