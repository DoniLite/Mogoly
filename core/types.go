package core

import (
	"net/http/httputil"
	"sync"
	"time"
)

type Server struct {
	ID              string    `json:"_" yaml:"_"`
	Name            string    `json:"name,omitempty" yaml:"name,omitempty"`
	Protocol        string    `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Host            string    `json:"host,omitempty" yaml:"host,omitempty"`
	Port            int       `json:"port,omitempty" yaml:"port,omitempty"`
	URL             string    `json:"url,omitempty" yaml:"url,omitempty"`
	IsHealthy       bool      `json:"is_healthy,omitempty" yaml:"is_healthy,omitempty"`
	LastHealthCheck time.Time `json:"__" yaml:"__"`
}

type ServerPool struct {
	servers map[string]*Server
	mu      sync.Mutex
}

type RoundRobinBalancer struct {
	pool *ServerPool // Reference to the server pooling object
	mu   sync.Mutex
	idx  int // The last selected server index
}

type BalancerStrategy interface {
	GetNextServer() (*Server, error)
}

type LoadBalancer struct {
	strategy BalancerStrategy
	proxy    *httputil.ReverseProxy
}

type ProxyServer struct {
	Host       string `json:"host" yaml:"host"`
	ListenPort string `json:"listen_port" yaml:"listen_port"`
}

type Config struct {
	Proxy   ProxyServer `json:"proxy" yaml:"proxy"`
	Servers []Server    `json:"server" yaml:"server"`
}
