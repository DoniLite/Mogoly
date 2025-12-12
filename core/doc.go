// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
# Mogoly Core

The core package is the main load balancing engine of the Mogoly software.
It provides comprehensive functionality for:

  - Server Pooling: Manage backend server pools with automatic health monitoring
  - Load Balancing: Round-robin distribution with automatic failover
  - Reverse Proxy: HTTP/HTTPS request forwarding with proper header handling
  - Health Checking: Periodic health checks with configurable intervals
  - Middleware System: Extensible middleware chain for rate limiting, logging, and more
  - Configuration Management: YAML/JSON config file support with hot-reloading
  - TLS/SSL Support: Automatic certificate management via CertMagic
  - DNS Support: Custom DNS resolver for local domain resolution
  - Dynamic Routing: Host-based routing with multiple server pools

# Architecture

The core package implements a sophisticated load balancing architecture:

	┌─────────────┐
	│   Client    │
	└──────┬──────┘
	       │
	       ▼
	┌─────────────────┐
	│     Router      │  (Host-based routing)
	└────────┬────────┘
	         │
	    ┌────┴────┐
	    ▼         ▼
	┌────────┐ ┌────────┐
	│Server 1│ │Server 2│  (Each can be a load balancer)
	└───┬────┘ └───┬────┘
	    │          │
	    ▼          ▼
	Backend    Backend
	 Pools      Pools

# Quick Start

## Basic Configuration

Create a configuration file (config.yaml):

	server:
	  - name: api-gateway
	    protocol: http
	    host: localhost
	    port: 3000
	    balance:
	      - name: backend-1
	        url: http://localhost:8081
	      - name: backend-2
	        url: http://localhost:8082
	    middlewares:
	      - name: ratelimit
	        config:
	          request_per_minute: 100
	          limit_window: 60s

	health_check_interval: 30
	stream: true

## Loading and Running

	package main

	import (
	    "log"
	    "github.com/DoniLite/Mogoly/core"
	)

	func main() {
	    // Load configuration
	    content, _ := core.LoadConfigFile("config.yaml")
	    format, _ := core.DiscoverConfigFormat("config.yaml")
	    config, _ := core.ParseConfig(content, format)

	    // Build router
	    core.BuildRouter(config)

	    // Start HTTP server
	    server := core.ServeHTTP(":8080")
	    defer server.Close()

	    // Keep running
	    select {}
	}

# Server Management

## Creating Servers Programmatically

	server := &core.Server{
	    Name:     "backend-1",
	    Protocol: "http",
	    Host:     "localhost",
	    Port:     8080,
	}

	// Initialize reverse proxy
	err := server.UpgradeProxy()
	if err != nil {
	    log.Fatal(err)
	}

	// Add to balancing pool
	backend := &core.Server{
	    Name: "backend-2",
	    URL:  "http://localhost:8081",
	}
	server.AddNewBalancingServer(backend)

	// Remove from pool
	server.DelBalancingServer("backend-2")

## Health Checking

	// Check all servers in pool
	status, err := server.CheckHealthAll()
	if err != nil {
	    log.Fatal(err)
	}

	for _, s := range status.Pass {
	    log.Printf("✓ %s is healthy", s.Name)
	}

	for _, s := range status.Fail {
	    log.Printf("✗ %s is down", s.Name)
	}

	// Check specific server
	serverStatus, _ := server.CheckHealthAny("backend-1")

	// Check self
	selfStatus, _ := server.CheckHealthSelf()

## Load Balancing

The core package uses a round-robin strategy with automatic failover:

	// Get next healthy server (automatically skips unhealthy servers)
	nextServer, err := server.GetNextServer()
	if err != nil {
	    log.Fatal("No healthy servers available")
	}

	// The server can be used as an HTTP handler
	http.Handle("/", server)

# Middleware System

## Built-in Middlewares

### Rate Limiter

Limits the number of requests per IP address:

	config := core.RateLimitMiddlewareConfig{
	    ReqPerMinute: 60,
	    LimitWindow:  time.Minute,
	}

	middleware := core.RateLimiterMiddleware(config)

## Chaining Middlewares

	handler := core.ChainMiddleware(
	    baseHandler,
	    authMiddleware,
	    rateLimitMiddleware,
	    loggingMiddleware,
	)

## Creating Custom Middlewares

	type CustomConfig struct {
	    Value string
	}

	func CustomMiddleware(config any) func(next http.Handler) http.Handler {
	    cfg := config.(CustomConfig)

	    return func(next http.Handler) http.Handler {
	        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	            // Pre-processing
	            log.Printf("Custom: %s", cfg.Value)

	            // Call next handler
	            next.ServeHTTP(w, r)

	            // Post-processing
	        })
	    }
	}

	// Register middleware
	core.MiddlewaresList["custom"] = core.MiddlewareSets{
	    Fn:   CustomMiddleware,
	    Conf: CustomConfig{Value: "example"},
	}

# Configuration Management

## Configuration File Formats

Supports both YAML and JSON formats:

### YAML Configuration

	server:
	  - name: api-gateway
	    protocol: http
	    host: localhost
	    port: 3000
	    balance:
	      - name: backend-1
	        protocol: http
	        host: localhost
	        port: 8081
	      - name: backend-2
	        url: http://localhost:8082
	    middlewares:
	      - name: ratelimit
	        config:
	          request_per_minute: 100
	          limit_window: 60s

	health_check_interval: 30
	log_output: /var/log/mogoly.log
	stream: true

### JSON Configuration

	{
	  "server": [
	    {
	      "name": "api-gateway",
	      "protocol": "http",
	      "host": "localhost",
	      "port": 3000,
	      "balance": [
	        {
	          "name": "backend-1",
	          "url": "http://localhost:8081"
	        }
	      ],
	      "middlewares": [
	        {
	          "name": "ratelimit",
	          "config": {
	            "request_per_minute": 100
	          }
	        }
	      ]
	    }
	  ],
	  "health_check_interval": 30,
	  "stream": true
	}

## Loading Configuration

	// Load config file
	content, err := core.LoadConfigFile("config.yaml")
	if err != nil {
	    log.Fatal(err)
	}

	// Auto-detect format from extension
	format, err := core.DiscoverConfigFormat("config.yaml")
	if err != nil {
	    log.Fatal(err)
	}

	// Parse configuration
	config, err := core.ParseConfig(content, format)
	if err != nil {
	    log.Fatal(err)
	}

# TLS/SSL Support

## HTTPS Server with Certificate Manager

	package main

	import (
	    "github.com/DoniLite/Mogoly/core"
	)

	func main() {
	    // Create certificate manager
	    certManager := core.NewCertManager(
	        "your-email@example.com",
	        []string{"example.com", "*.example.com"},
	    )

	    // Build router
	    core.BuildRouter(config)

	    // Start HTTPS server
	    httpsServer := core.ServeHTTPS(":443", certManager)
	    defer httpsServer.Close()

	    // Optionally start HTTP server for redirects
	    httpServer := core.ServeHTTP(":80")
	    defer httpServer.Close()

	    select {}
	}

The certificate manager automatically:
  - Obtains certificates from Let's Encrypt
  - Renews certificates before expiration
  - Handles ACME challenges
  - Caches certificates on disk

# Router System

## Dynamic Routing

The router maps hostnames to server handlers:

	// Build router from config
	core.BuildRouter(config)

	// Get current router
	router := core.GetRouter()

	// Add server dynamically
	newServer := &core.Server{
	    Name: "new-api",
	    URL:  "http://localhost:9000",
	}
	router.AddServer(newServer)

	// Remove server
	router.RemoveServer(newServer)

## Host-Based Routing

Requests are routed based on the Host header:

	example.com     -> Server "example.com"
	api.example.com -> Server "api.example.com"
	*.example.com   -> Server "wildcard.example.com" (if configured)

# Best Practices

1. **Use Health Checks**: Configure appropriate health check intervals (30-60 seconds recommended)

2. **Configure Timeouts**: Set reasonable HTTP client timeouts for health checks (3-5 seconds)

3. **Initialize Proxies**: Always call UpgradeProxy() on servers before using them as load balancers

4. **Middleware Order**: Chain middlewares in the correct order:
  - Authentication first
  - Rate limiting second
  - Logging last

5. **TLS in Production**: Use CertMagic for automatic certificate management

6. **Monitor Health**: Set up alerts for when servers become unhealthy

7. **Graceful Shutdown**: Properly close HTTP servers on application exit

8. **Configuration Validation**: Validate configuration files before deploying

9. **Resource Cleanup**: Always defer Close() on servers and managers

10. **Load Distribution**: Use multiple backend servers for better distribution

# Types Reference

## Server

	type Server struct {
	    ID               string       // Server ID
	    Name             string       // Server name (required)
	    Protocol         string       // "http" or "https"
	    Host             string       // Hostname or IP
	    Port             int          // Port number
	    URL              string       // Full URL (alternative to host+port)
	    IsHealthy        bool         // Health status
	    BalancingServers []*Server    // Backend pool for load balancing
	    Middlewares      []Middleware // Applied middlewares
	    LastHealthCheck  *time.Time   // Last health check timestamp
	}

## Config

	type Config struct {
	    Servers             []*Server              // Server definitions
	    HealthCheckInterval int                    // Interval in seconds
	    LogOutput           string                 // Log file path
	    Stream              bool                   // Enable log streaming
	    Services            []*cloud.ServiceConfig // Cloud services
	}

## HealthCheckStatus

	type HealthCheckStatus struct {
	    Pass      []ServerStatus // Healthy servers
	    Fail      []ServerStatus // Unhealthy servers
	    CheckTime time.Time      // Check timestamp
	    Duration  time.Duration  // Check duration
	}

# Integration with Cloud Package

The core package can automatically provision database services:

	config := &core.Config{
	    Servers: []*core.Server{
	        {
	            Name: "api",
	            URL:  "http://localhost:3000",
	        },
	    },
	    Services: []*cloud.ServiceConfig{
	        {
	            Type:     cloud.PostgreSQL,
	            Name:     "app-database",
	            Username: "admin",
	            Password: "secure",
	        },
	    },
	}

The database service will be automatically created and managed alongside your load balancer.

# Performance Considerations

- **Connection Pooling**: The reverse proxy reuses connections to backend servers
- **Health Check Optimization**: Failed servers are temporarily marked unhealthy
- **Middleware Efficiency**: Keep middleware processing lightweight
- **Router Locks**: The router uses RWMutex for efficient concurrent access
- **Memory Usage**: Each server maintains its own proxy and connection pool

# Troubleshooting

## No healthy servers available

- Check if backend servers are running
- Verify health check URLs are accessible
- Review health check interval settings
- Check server logs for connection errors

## Certificate errors

- Verify domain DNS records point to your server
- Ensure ports 80 and 443 are accessible
- Check email address is valid for ACME
- Review certificate manager logs

## High memory usage

- Reduce number of concurrent connections
- Adjust health check intervals
- Monitor middleware memory consumption
- Check for connection leaks in custom handlers

# See Also

  - Package sync: Real-time communication between components
  - Package cloud: Docker service management and provisioning
*/
package core
