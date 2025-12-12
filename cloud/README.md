# @Mogoly/cloud

This package provides a high-level API for Docker service management, enabling programmatic creation and deployment of database services with custom domain assignment and HTTPS support via Traefik integration.

## Features

- **Docker Service Management**: Create, list, and delete database service instances
- **Supported Databases**: PostgreSQL, MySQL, MongoDB, Redis, MariaDB
- **Custom Domain Support**: Assign custom domains to services with automatic SSL/TLS via Let's Encrypt
- **Traefik Integration**: Automated reverse proxy setup with ACME support
- **Volume Management**: Persistent storage configuration for database containers
- **Automatic Port Allocation**: Smart port management for service exposure

## Prerequisites

- [Docker](https://www.docker.com/) must be installed and running
- Docker daemon must be accessible
- For custom domains: DNS records must point to your server

## Installation

```bash
go get github.com/DoniLite/Mogoly/cloud
```

## Quick Start

### Basic Usage

```go
package main

import (
    "log"
    "github.com/DoniLite/Mogoly/cloud"
)

func main() {
    // Create cloud manager
    manager, err := cloud.NewCloudManager()
    if err != nil {
        log.Fatal(err)
    }
    defer manager.Close()

    // Configure a PostgreSQL service
    config := cloud.ServiceConfig{
        Type:         cloud.PostgreSQL,
        Name:         "my-postgres-db",
        Username:     "admin",
        Password:     "securepassword",
        DatabaseName: "mydb",
        Version:      "14",
        Volumes: []cloud.VolumeMount{
            {
                Name:          "postgres-data",
                ContainerPath: "/var/lib/postgresql/data",
            },
        },
    }

    // Create the service instance
    instance, err := manager.CreateInstance(config)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Service created: %s on port %d", instance.Name, instance.ExternalPort)
}
```

## API Reference

### CloudManager

The main entry point for managing Docker services.

#### Creating the Manager

```go
manager, err := cloud.NewCloudManager()
if err != nil {
    log.Fatal(err)
}
defer manager.Close()
```

#### Methods

##### CreateInstance

Create a new database service instance.

```go
func (m *CloudManager) CreateInstance(config ServiceConfig) (*ServiceInstance, error)
```

**Parameters:**
- `config`: Service configuration (type, name, credentials, volumes, domain, etc.)

**Returns:**
- `*ServiceInstance`: The created service instance
- `error`: Any error that occurred during creation

**Example:**

```go
config := cloud.ServiceConfig{
    Type:     cloud.MySQL,
    Name:     "my-mysql",
    Username: "root",
    Password: "rootpass",
    Version:  "8.0",
}

instance, err := manager.CreateInstance(config)
```

##### GetInstance

Retrieve a service instance by ID.

```go
func (m *CloudManager) GetInstance(id string) (*ServiceInstance, bool)
```

##### ListInstances

Get all service instances.

```go
func (m *CloudManager) ListInstances() []*ServiceInstance
```

##### DeleteInstance

Delete a service instance (removes container and volumes).

```go
func (m *CloudManager) DeleteInstance(id string) error
```

**Warning:** This will permanently delete the container and its data.

##### CreateTraefikBundle

Create and start the Traefik reverse proxy with ACME support.

```go
func (m *CloudManager) CreateTraefikBundle(acmeEmail string) (string, error)
```

**Parameters:**
- `acmeEmail`: Email address for Let's Encrypt certificate registration

**Returns:**
- `string`: Container ID of the Traefik instance
- `error`: Any error that occurred

**Note:** This must be called before creating services with custom domains.

##### RecreateWithDomain

Recreate an existing service with a custom domain.

```go
func (m *CloudManager) RecreateWithDomain(id string, domain *DomainConfig) (*ServiceInstance, error)
```

### Types

#### ServiceType

```go
type ServiceType string

const (
    PostgreSQL ServiceType = "postgresql"
    MySQL      ServiceType = "mysql"
    MongoDB    ServiceType = "mongodb"
    Redis      ServiceType = "redis"
    MariaDB    ServiceType = "mariadb"
)
```

#### ServiceConfig

```go
type ServiceConfig struct {
    Type            ServiceType       // Database type
    Name            string            // Service name (required)
    Username        string            // Database username
    Password        string            // Database password
    DatabaseName    string            // Database name
    Version         string            // Specific version tag (e.g., "14", "8.0")
    UseLatestVersion bool             // Use latest version instead
    Variables       map[string]string // Additional environment variables
    ExposedPorts    []string          // Additional ports to expose
    Volumes         []VolumeMount     // Volume configurations
    Domain          *DomainConfig     // Custom domain configuration
}
```

#### ServiceInstance

```go
type ServiceInstance struct {
    ID           string      // Unique identifier
    Name         string      // Service name
    Type         ServiceType // Database type
    ContainerID  string      // Docker container ID
    Port         int         // Internal port
    Username     string      // Database username
    Password     string      // Database password
    DatabaseName string      // Database name
    Status       string      // Current status
    CreatedAt    time.Time   // Creation timestamp
    ExternalPort int         // Exposed external port
    Domain       string      // Custom domain
    VolumeNames  []string    // Attached volume names
}
```

#### DomainConfig

```go
type DomainConfig struct {
    Domain       string   // e.g., "db.example.com"
    CertResolver string   // Default: "letsencrypt"
    EntryPoint   string   // Default: "websecure" (port 443)
    AllowedCIDRs []string // IP allowlist (optional)
}
```

#### VolumeMount

```go
type VolumeMount struct {
    Name          string            // Docker volume name
    ContainerPath string            // Mount path in container
    ReadOnly      bool              // Read-only mount
    Driver        string            // Volume driver (e.g., "local")
    DriverOpts    map[string]string // Driver-specific options
}
```

## Advanced Usage

### Using Custom Domains with HTTPS

To use custom domains with automatic SSL/TLS certificates:

**Step 1: Create Traefik Bundle**

```go
traefikID, err := manager.CreateTraefikBundle("your-email@example.com")
if err != nil {
    log.Fatal("Failed to create Traefik:", err)
}
log.Printf("Traefik started with ID: %s", traefikID)
```

**Step 2: Create Service with Domain**

```go
config := cloud.ServiceConfig{
    Type:     cloud.PostgreSQL,
    Name:     "production-db",
    Username: "admin",
    Password: "securepass",
    Domain: &cloud.DomainConfig{
        Domain:       "db.example.com",
        CertResolver: "letsencrypt",
        EntryPoint:   "websecure",
    },
    Volumes: []cloud.VolumeMount{
        {
            Name:          "prod-db-data",
            ContainerPath: "/var/lib/postgresql/data",
        },
    },
}

instance, err := manager.CreateInstance(config)
if err != nil {
    log.Fatal(err)
}

log.Printf("Production database available at: https://%s", instance.Domain)
```

### Custom Volume Configuration

```go
config := cloud.ServiceConfig{
    Type: cloud.MongoDB,
    Name: "mongo-cluster",
    Volumes: []cloud.VolumeMount{
        {
            Name:          "mongo-data",
            ContainerPath: "/data/db",
            Driver:        "local",
            DriverOpts: map[string]string{
                "type":   "nfs",
                "o":      "addr=10.0.0.1,rw",
                "device": ":/path/to/dir",
            },
        },
        {
            Name:          "mongo-config",
            ContainerPath: "/data/configdb",
            ReadOnly:      false,
        },
    },
}

instance, err := manager.CreateInstance(config)
```

### Environment Variables

```go
config := cloud.ServiceConfig{
    Type: cloud.MySQL,
    Name: "configured-mysql",
    Variables: map[string]string{
        "MYSQL_MAX_CONNECTIONS":    "500",
        "MYSQL_INNODB_BUFFER_POOL": "2G",
    },
}
```

## Best Practices

1. **Always use volumes**: Configure persistent volumes to prevent data loss
2. **Secure credentials**: Use strong passwords and consider using secrets management
3. **Close the manager**: Always defer `manager.Close()` to clean up resources
4. **Create Traefik first**: If using custom domains, create the Traefik bundle before services
5. **Use specific versions**: Specify database versions instead of using "latest" for production
6. **Monitor logs**: Check Docker logs if services fail to start
7. **DNS configuration**: Ensure DNS records point to your server before creating domain-enabled services

## Troubleshooting

### Service creation fails

- Ensure Docker daemon is running: `docker ps`
- Check if the port is already in use
- Verify Docker network exists: `docker network ls`
- Check Docker logs: `docker logs <container-id>`

### Custom domain SSL issues

- Verify DNS records point to your server
- Ensure ports 80 and 443 are accessible
- Check Traefik logs: `docker logs <traefik-container-id>`
- Verify email address is valid for ACME registration

### Volume mount issues

- Ensure the volume name is unique
- Check volume driver is available: `docker volume ls`
- Verify container path is valid for the database type

## Examples

See the [integration tests](file:///home/doni/Documents/Projects/Perso/Mogoly/cloud/integration_docker_test.go) for more usage examples.

## License

MIT License - see [LICENSE](file:///home/doni/Documents/Projects/Perso/Mogoly/cloud/LICENSE) for details.

## Contributing

This package is part of the [Mogoly](https://github.com/DoniLite/Mogoly) project.
