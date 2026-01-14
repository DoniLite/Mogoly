# Domain Management - Complete Guide

## Overview

Mogoly's domain management system allows you to create custom domains with HTTPS for both local development and production. This guide shows you exactly how to bind domains to your local services.

---

## Quick Start: Bind Domain to Local Server

### Scenario: You have a service running on `http://localhost:3000`

**Goal**: Access it via `https://myapp.local`

**Steps**:

```bash
# 1. Start your local service (e.g., React app, API server)
npm start  # Running on http://localhost:3000

# 2. Start Mogoly daemon
mogoly daemon start

# 3. Create a local domain (requires sudo for /etc/hosts)
sudo mogoly domain add myapp.local --local

# 4. Create a load balancer
mogoly lb create --name myapp --port 443

# 5. Add your local service as a backend
mogoly lb add-backend myapp --url http://localhost:3000 --name frontend

# 6. Access your service
curl https://myapp.local
# → Routes to http://localhost:3000
```

**What happened**:
1. Created self-signed certificate for `myapp.local`
2. Added `/etc/hosts` entry: `127.0.0.1 myapp.local`
3. Load balancer listens on HTTPS port 443
4. Requests to `https://myapp.local` → proxied to `http://localhost:3000`

---

## Usage Patterns

### Pattern 1: Single Service (Direct)

**When**: Simple setup, one service, no load balancing needed

```bash
# Create domain
sudo mogoly domain add api.local --local

# Your service runs on any port
node server.js  # http://localhost:8080

# Access it
https://api.local:8080  # Direct to your service
```

**How it works**: Domain resolves to `127.0.0.1`, your service handles its own port

**Use Case**: Quick local development, testing SSL locally

---

### Pattern 2: Load Balancer Proxy (Recommended)

**When**: Want standard HTTPS port (443), load balancing, or routing

```bash
# Create domain
sudo mogoly domain add app.local --local

# Create load balancer on standard HTTPS port
mogoly lb create --name app --port 443

# Add backend
mogoly lb add-backend app --url http://localhost:3000

# Access without port number
https://app.local  # Clean URL!
```

**How it works**: Load balancer proxies HTTPS:443 → HTTP:3000

**Use Case**: Production-like local environment

---

### Pattern 3: Multiple Services, Multiple Domains

**When**: Microservices architecture

```bash
# Frontend
sudo mogoly domain add frontend.local --local
mogoly lb create --name frontend --port 443
mogoly lb add-backend frontend --url http://localhost:3000

# API
sudo mogoly domain add api.local --local
mogoly lb create --name api --port 8443  # Different port!
mogoly lb add-backend api --url http://localhost:8080

# Database
sudo mogoly domain add db.local --local
# No load balancer needed, direct access:
psql postgresql://127.0.0.1:5432  # or db.local:5432
```

**Access**:
- Frontend: `https://frontend.local`
- API: `https://api.local:8443`
- Database: `db.local:5432`

---

### Pattern 4: Multiple Backends (Load Balancing)

**When**: Testing load balancing locally

```bash
# Create domain
sudo mogoly domain add myapp.local --local

# Create load balancer
mogoly lb create --name myapp --port 443

# Add multiple backends
mogoly lb add-backend myapp --url http://localhost:3000 --name backend1
mogoly lb add-backend myapp --url http://localhost:3001 --name backend2
mogoly lb add-backend myapp --url http://localhost:3002 --name backend3

# Requests are round-robin distributed
curl https://myapp.local  # → backend1
curl https://myapp.local  # → backend2
curl https://myapp.local  # → backend3
```

---

## Complete Real-World Examples

### Example 1: React + Node.js App

```bash
# Start services
cd frontend && npm start  # http://localhost:3000
cd backend && npm start   # http://localhost:8080

# Setup domains
sudo mogoly domain add myapp.local --local
sudo mogoly domain add api.myapp.local --local

# Create load balancers
mogoly daemon start
mogoly lb create --name frontend --port 443
mogoly lb create --name api --port 443

# Add backends
mogoly lb add-backend frontend --url http://localhost:3000
mogoly lb add-backend api --url http://localhost:8080

# Configure frontend to call API
# In React app, use: https://api.myapp.local

# Access
open https://myapp.local
```

---

### Example 2: Docker Containers

```bash
# Start containers
docker run -d -p 8080:80 nginx   # Container 1
docker run -d -p 8081:80 nginx   # Container 2

# Setup domain
sudo mogoly domain add nginx.local --local

# Create load balancer
mogoly lb create --name nginx --port 443

# Add containers as backends
mogoly lb add-backend nginx --url http://localhost:8080 --name container1
mogoly lb add-backend nginx --url http://localhost:8081 --name container2

# Access
curl https://nginx.local
```

---

### Example 3: Full Stack with Database

```bash
# Start everything
docker run -d -p 5432:5432 postgres    # Database
cd api && npm start                     # http://localhost:8080
cd frontend && npm start                # http://localhost:3000

# Setup domains
sudo mogoly domain add app.local --local
sudo mogoly domain add api.local --local
sudo mogoly domain add db.local --local

# Create load balancers for web services
mogoly lb create --name frontend --port 443
mogoly lb create --name api --port 443

# Add backends
mogoly lb add-backend frontend --url http://localhost:3000
mogoly lb add-backend api --url http://localhost:8080

# Database uses direct connection (no LB needed)
# Connection string: postgresql://db.local:5432/mydb

# Access
https://app.local           # Frontend
https://api.local           # API
psql -h db.local -p 5432   # Database
```

---

## Port Handling Deep Dive

### Without Load Balancer

```
Domain: api.local
Resolves to: 127.0.0.1

Your services:
- Service A: http://localhost:3000
- Service B: http://localhost:8080

Access via:
- https://api.local:3000 → Service A
- https://api.local:8080 → Service B
```

### With Load Balancer

```
Domain: api.local
Load Balancer: port 443

Backend: http://localhost:3000

Access via:
- https://api.local (port 443 implied) → proxied toLB → Service
```

### Custom Ports

```bash
# Load balancer on non-standard port
mogoly lb create --name myapp --port 8443

# Access
https://myapp.local:8443
```

---

## Independent Usage (No Load Balancer)

If you just want HTTPS certificates for your own server:

```bash
# Create domain
sudo mogoly domain add myservice.local --local

# Get certificate paths
ls ~/.mogoly/certs/
# myservice.local.crt
# myservice.local.key

# Use in your Node.js server
const https = require('https');
const fs = require('fs');

const options = {
  key: fs.readFileSync('/home/user/.mogoly/certs/myservice.local.key'),
  cert: fs.readFileSync('/home/user/.mogoly/certs/myservice.local.crt')
};

https.createServer(options, app).listen(443);

# Access
https://myservice.local
```

---

## Production Deployment

Same workflow, but with real domains:

```bash
# Prerequisites:
# 1. DNS points to your server (api.myapp.com → your_server_ip)
# 2. Ports 80/443 open

# Create domain (no sudo needed, no /etc/hosts)
mogoly domain add api.myapp.com --auto-ssl

# Create load balancer
mogoly lb create --name production --port 443

# Add backends
mogoly lb add-backend production --url http://10.0.1.10:8080 --name server1
mogoly lb add-backend production --url http://10.0.1.11:8080 --name server2

# Let's Encrypt automatically obtains certificate
# Auto-renewal before expiry

# Access
curl https://api.myapp.com
```

---

## Troubleshooting

### Domain not resolving

```bash
# Check hosts file
cat /etc/hosts | grep myapp.local
# Should show: 127.0.0.1 myapp.local # Managed by Mogoly

# Flush DNS
sudo dscacheutil -flushcache  # macOS
sudo systemctl restart systemd-resolved  # Linux
ipconfig /flushdns  # Windows

# Test resolution
ping myapp.local
# Should ping 127.0.0.1
```

### Certificate warnings

```bash
# For local development, this is normal
# Option 1: Click "Advanced" → "Proceed" in browser

# Option 2: Trust certificate permanently
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain \
  ~/.mogoly/certs/myapp.local.crt

# Option 3: Use curl with -k flag
curl -k https://myapp.local
```

### Load balancer not starting

```bash
# Check if port is already in use
sudo lsof -i :443

# Check daemon status
mogoly daemon status

# View logs
mogoly daemon logs --tail 50
```

### Permission denied

```bash
# Domain creation needs sudo for /etc/hosts
sudo mogoly domain add myapp.local --local

# Alternative: Run daemon as sudo (not recommended)
```

---

## Best Practices

1. **Use .local for development**: Avoids conflicts with real domains
2. **Standard ports for production feel**: Use port 443 for HTTPS
3. **One domain per service**: Cleaner organization
4. **Load balancer for routing**: Even with one backend, easier to manage
5. **Test locally before production**: Same domain setup, different backend URLs

---

## Summary

| Scenario | Command | Result |
|----------|---------|--------|
| Simple local dev | `sudo mogoly domain add app.local --local` | Access service by domain, any port |
| With load balancer | `+ mogoly lb create` + `add-backend` | HTTPS:443 → your service |
| Multiple services | Create domain per service | Each service has own HTTPS domain |
| Production | `--auto-ssl` instead of `--local` | Let's Encrypt certificate |

Happy developing with HTTPS! 🔒
