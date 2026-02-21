package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/DoniLite/Mogoly/cloud"
	"github.com/DoniLite/Mogoly/core/config"
	"github.com/DoniLite/Mogoly/core/events"
	"github.com/DoniLite/Mogoly/core/server"
)

const (
	ROUTER_CONFIG_FILE string = "router.yaml"
)

// Servers

func (rs *RouterState) AddServer(server *server.Server) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if _, exists := rs.serverMap[strings.ToLower(server.Name)]; exists {
		return
	}
	rs.serverMap[strings.ToLower(server.Name)] = server
	rs.httpServerMap[strings.ToLower(server.Name)] = CreateSingleHttpServer(server)
}

func (rs *RouterState) RemoveServer(server *server.Server) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if _, exists := rs.serverMap[strings.ToLower(server.Name)]; !exists {
		return
	}
	delete(rs.httpServerMap, strings.ToLower(server.Name))
	delete(rs.serverMap, strings.ToLower(server.Name))
}

func (rs *RouterState) GetServer(name string) (*server.Server, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	if _, exists := rs.serverMap[strings.ToLower(name)]; !exists {
		return nil, fmt.Errorf("server %s not found", name)
	}
	return rs.serverMap[strings.ToLower(name)], nil
}

func (rs *RouterState) GetHandler(name string) (http.Handler, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	if _, exists := rs.httpServerMap[strings.ToLower(name)]; !exists {
		return nil, fmt.Errorf("handler %s not found", name)
	}
	return rs.httpServerMap[strings.ToLower(name)], nil
}

func (rs *RouterState) ListServers() []*server.Server {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	servers := make([]*server.Server, 0, len(rs.serverMap))
	for _, server := range rs.serverMap {
		servers = append(servers, server)
	}
	return servers
}

// Services

func (rs *RouterState) AddService(service *cloud.ServiceConfig) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if _, exists := rs.cloudMap[strings.ToLower(service.Name)]; exists {
		return
	}
	currentService, err := rs.serviceManager.CreateInstance(*service)
	if err != nil {
		events.Logf(events.LOG_ERROR, "[ROUTER]: Error while creating service instance for %s: %v", service.Name, err)
		return
	}
	rs.cloudMap[strings.ToLower(service.Name)] = service
	rs.cloudServiceInstanceMap[strings.ToLower(service.Name)] = currentService
}

func (rs *RouterState) RemoveService(service *cloud.ServiceConfig) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if _, exists := rs.cloudMap[strings.ToLower(service.Name)]; !exists {
		return
	}
	serviceInstance, exists := rs.cloudServiceInstanceMap[strings.ToLower(service.Name)]
	if !exists {
		return
	}
	if err := rs.serviceManager.DeleteInstance(serviceInstance.ID); err != nil {
		events.Logf(events.LOG_ERROR, "[ROUTER]: Error while deleting service instance for %s: %v", service.Name, err)
		return
	}
	delete(rs.cloudServiceInstanceMap, strings.ToLower(service.Name))
	delete(rs.cloudMap, strings.ToLower(service.Name))
}

func (rs *RouterState) GetService(name string) (*cloud.ServiceConfig, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	if _, exists := rs.cloudMap[strings.ToLower(name)]; !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}
	return rs.cloudMap[strings.ToLower(name)], nil
}

func (rs *RouterState) GetServiceInstance(name string) (*cloud.ServiceInstance, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	if _, exists := rs.cloudServiceInstanceMap[strings.ToLower(name)]; !exists {
		return nil, fmt.Errorf("service instance %s not found", name)
	}
	return rs.cloudServiceInstanceMap[strings.ToLower(name)], nil
}

// Config

func (cf *Config) BuildVars() {
	for k, v := range cf.Variables {
		config.SetEnv(k, v)
	}
}
