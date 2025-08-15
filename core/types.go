// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"net/http/httputil"
	"sync"
	"time"
)

type LogType = int

const (
	LOG_INFO LogType = iota
	LOG_DEBUG
	LOG_ERROR
)

type Server struct {
	ID              string // THe server ID based on its registration order
	Name            string `json:"name,omitempty" yaml:"name,omitempty"`             // The server name
	Protocol        string `json:"protocol,omitempty" yaml:"protocol,omitempty"`     // The protocol for the server this field can be `http` or `https`
	Host            string `json:"host,omitempty" yaml:"host,omitempty"`             // The server host
	Port            int    `json:"port,omitempty" yaml:"port,omitempty"`             // The port on which the server is running
	URL             string `json:"url,omitempty" yaml:"url,omitempty"`               // If this field is provided the URL will be used for request forwarding
	IsHealthy       bool   `json:"is_healthy,omitempty" yaml:"is_healthy,omitempty"` // Specifying the server health check state
	LastHealthCheck time.Time
	Proxy           *httputil.ReverseProxy
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

type Logs struct {
	logType LogType
	message string
}

type LoadBalancer struct {
	strategy BalancerStrategy
	Logs     chan Logs
}

type ProxyServer struct {
	Name       string `json:"name" yaml:"name"`               // Specifying a proxy server name
	Host       string `json:"host" yaml:"host"`               // The host to start the server
	ListenPort string `json:"listen_port" yaml:"listen_port"` // The server Port
}

type Config struct {
	Proxy   ProxyServer `json:"proxy" yaml:"proxy"`   // The Proxy config specification
	Servers []Server    `json:"server" yaml:"server"` // The servers instances
}

// The result of a health checking process for a server
type ServerStatus struct {
	Name    string `json:"name" yaml:"name"`       // The server name
	Url     string `json:"url" yaml:"url"`         // HealthCheck url
	Healthy bool   `json:"healthy" yaml:"healthy"` // healthCheck status
}

type HealthCheckStatus struct {
	Pass      []ServerStatus `json:"pass" yaml:"pass"` // Array of successful HealthCheck result
	Fail      []ServerStatus `json:"fail" yaml:"fail"` // Array of failure HealthCheck Result
	CheckTime time.Time
	Duration  time.Duration
}
