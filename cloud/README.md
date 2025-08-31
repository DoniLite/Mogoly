# @Mogoly/cloud

This package provides a high level API for docker service building and custom domains assignment with `http/https`.  
The API is used as a dependency for the `@Mogoly/core` package. To use this package be sure that you've installed [docker](https://www.google.com/url?sa=t&source=web&rct=j&opi=89978449&url=https://www.docker.com/&ved=2ahUKEwisr7nPla6PAxXaTDABHSffPIkQFnoECBIQAQ&usg=AOvVaw3p9e1qPvdfjCrUwPYAhUlS) in your environment.

To activate `https` make sure to run the traefik service with the method:

```go
func (m *CloudManager) CreateTraefikBundle(acmeEmail string) (string, error)
```

Refer to the types to see the config to provide to activate `tls` for a specific service
> This feature is only available for defined database services
