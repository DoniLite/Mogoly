package server

import (
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

type Server struct {
	ID               string       // THe server ID based on its registration order
	Name             string       `json:"name,omitempty" yaml:"name,omitempty"`             // The server name
	Protocol         string       `json:"protocol,omitempty" yaml:"protocol,omitempty"`     // The protocol for the server this field can be `http` or `https`
	Host             string       `json:"host,omitempty" yaml:"host,omitempty"`             // The server host
	Port             int          `json:"port,omitempty" yaml:"port,omitempty"`             // The port on which the server is running
	URL              string       `json:"url,omitempty" yaml:"url,omitempty"`               // If this field is provided the URL will be used for request forwarding
	IsHealthy        bool         `json:"is_healthy,omitempty" yaml:"is_healthy,omitempty"` // Specifying the server health check state
	BalancingServers []*Server    `json:"balance,omitempty" yaml:"balance,omitempty"`       // If specified these servers will be used for load balancing request
	Middlewares      []Middleware `json:"middlewares,omitempty" yaml:"middlewares,omitempty"`
	LastHealthCheck  *time.Time
	proxy            *httputil.ReverseProxy
	mu               sync.Mutex
	idx              int
	ForceTLS         bool
}

type Middleware struct {
	Name   string `json:"name" yaml:"name"`
	Config any    `json:"config,omitempty" yaml:"config,omitempty"`
}

type ServerStrategy string

const (
	ServerStrategyRoundRobin ServerStrategy = "round_robin"
	ServerStrategyRandom     ServerStrategy = "random"
)

type BalancerStrategy interface {
	GetNextServer(strategy ServerStrategy) (*Server, error)
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

type MogolyMiddleware func(config any) func(next http.Handler) http.Handler

type MiddleWareName string

type MiddlewareSets map[MiddleWareName]struct {
	Fn   MogolyMiddleware
	Conf any
}

