package router

import (
	"fmt"
	"strings"

	"github.com/DoniLite/Mogoly/core/config"
	"github.com/DoniLite/Mogoly/core/server"
)

func (rs *RouterState) AddServer(server *server.Server) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if _, exists := rs.s[strings.ToLower(server.Name)]; exists {
		return
	}
	rs.s[strings.ToLower(server.Name)] = server
	rs.m[strings.ToLower(server.Name)] = CreateSingleHttpServer(server)
}

func (rs *RouterState) RemoveServer(server *server.Server) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if _, exists := rs.s[strings.ToLower(server.Name)]; !exists {
		return
	}
	delete(rs.m, strings.ToLower(server.Name))
	delete(rs.s, strings.ToLower(server.Name))
}

func (rs *RouterState) GetServer(name string) (*server.Server, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	if _, exists := rs.s[strings.ToLower(name)]; !exists {
		return nil, fmt.Errorf("server %s not found", name)
	}
	return rs.s[strings.ToLower(name)], nil
}

func (rs *RouterState) UpdateConfig(config *Config) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	for _, server := range config.Servers {
		rs.AddServer(server)
	}
}


// Config

func (cf *Config) BuildVars() {
	for k, v := range cf.Variables {
		config.SetEnv(k, v)
	}
}