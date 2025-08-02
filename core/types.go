package core

import (
	"sync"
	"time"
)

type Server struct {
	ID              string    `json:"id,omitempty" yaml:"id,omitempty"`
	Name            string    `json:"name,omitempty" yaml:"name,omitempty"`
	Protocol        string    `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Host            string    `json:"host,omitempty" yaml:"host,omitempty"`
	Port            int       `json:"port,omitempty" yaml:"port,omitempty"`
	URL             string    `json:"url,omitempty" yaml:"url,omitempty"`
	IsHealthy       bool      `json:"is_healthy,omitempty" yaml:"is_healthy,omitempty"`
	LastHealthCheck time.Time `json:"last_healthcheck,omitempty" yaml:"last_healthcheck,omitempty"`
}

type ServerPool struct {
	Servers map[string]*Server
	Mu      sync.Mutex
}

type RoundRobinBalancer struct {
	Pool *ServerPool // Reference to the server pooling object
	Mu   sync.Mutex
	Idx  int // The last selected server index
}
