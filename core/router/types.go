package router

import (
	"net/http"
	"sync"

	"github.com/DoniLite/Mogoly/cloud"
	"github.com/DoniLite/Mogoly/core/server"
)

type RouterState struct {
	mu           sync.RWMutex
	m            map[string]http.Handler // host -> backend
	s            map[string]*server.Server
	globalConfig *Config
}

type Config struct {
	Servers             []*server.Server       `json:"server" yaml:"server"` // The servers instances
	HealthCheckInterval int                    `json:"healthcheck_interval,omitempty" yaml:"healthcheck_interval,omitempty"`
	LogOutput           string                 `json:"log_output,omitempty" yaml:"log_output,omitempty"`
	Stream              bool                   `json:"stream,omitempty" yaml:"stream,omitempty"`
	Services            []*cloud.ServiceConfig `json:"services,omitempty" yaml:"services,omitempty"`
	Variables           map[string]string      `json:"variables,omitempty" yaml:"variables,omitempty"`
}
