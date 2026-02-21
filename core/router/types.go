package router

import (
	"net/http"
	"sync"

	"github.com/DoniLite/Mogoly/cloud"
	"github.com/DoniLite/Mogoly/core/server"
)

type RouterState struct {
	mu                      sync.RWMutex
	httpServerMap           map[string]http.Handler // host -> backend
	serverMap               map[string]*server.Server
	cloudMap                map[string]*cloud.ServiceConfig
	cloudServiceInstanceMap map[string]*cloud.ServiceInstance
	serviceManager          *cloud.CloudManager

	globalConfig *Config
}

type Config struct {
	Servers   []*server.Server       `json:"server" yaml:"server"` // The servers instances
	Services  []*cloud.ServiceConfig `json:"services,omitempty" yaml:"services,omitempty"`
	Variables map[string]string      `json:"variables,omitempty" yaml:"variables,omitempty"`
}
