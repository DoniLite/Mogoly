package router

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/DoniLite/Mogoly/cloud"
	"github.com/DoniLite/Mogoly/core/config"
	"github.com/DoniLite/Mogoly/core/events"
	"github.com/DoniLite/Mogoly/core/server"
	"gopkg.in/yaml.v3"
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

	err := rs.globalConfig.PersistConfig()
	if err != nil {
		events.Logf(events.LOG_ERROR, "[ROUTER]: Error while persisting router config: %v", err)
	}
}

func (rs *RouterState) RemoveServer(server *server.Server) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if _, exists := rs.serverMap[strings.ToLower(server.Name)]; !exists {
		return
	}
	delete(rs.httpServerMap, strings.ToLower(server.Name))
	delete(rs.serverMap, strings.ToLower(server.Name))

	err := rs.globalConfig.PersistConfig()
	if err != nil {
		events.Logf(events.LOG_ERROR, "[ROUTER]: Error while persisting router config: %v", err)
	}
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
	events.Logf(events.LOG_INFO, "[ROUTER]: Creating service instance for %s", service.Name)

	currentService, err := rs.serviceManager.CreateInstance(*service)
	if err != nil {
		events.Logf(events.LOG_ERROR, "[ROUTER]: Error while creating service instance for %s: %v", service.Name, err)
		return
	}
	rs.cloudMap[strings.ToLower(service.Name)] = service
	rs.cloudServiceInstanceMap[strings.ToLower(service.Name)] = currentService

	events.Logf(events.LOG_INFO, "[ROUTER]: Service instance for %s created successfully", service.Name)

	err = rs.globalConfig.PersistConfig()
	if err != nil {
		events.Logf(events.LOG_ERROR, "[ROUTER]: Error while persisting router config: %v", err)
	}
}

func (rs *RouterState) RemoveService(service *cloud.ServiceConfig) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if _, exists := rs.cloudMap[strings.ToLower(service.Name)]; !exists {
		return
	}

	events.Logf(events.LOG_INFO, "[ROUTER]: Removing service instance for %s", service.Name)

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

	events.Logf(events.LOG_INFO, "[ROUTER]: Service instance for %s deleted successfully", service.Name)

	err := rs.globalConfig.PersistConfig()
	if err != nil {
		events.Logf(events.LOG_ERROR, "[ROUTER]: Error while persisting router config: %v", err)
	}
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

func (rs *RouterState) RecreateInstanceWithDomain(id string, domain *cloud.DomainConfig) error {
	instance, isAvailable := rs.serviceManager.GetInstance(id)
	if !isAvailable {
		return fmt.Errorf("service instance %s not found", id)
	}

	if _, exists := rs.cloudServiceInstanceMap[strings.ToLower(instance.Name)]; !exists {
		return fmt.Errorf("service instance %s not found", instance.Name)
	}

	events.Logf(events.LOG_INFO, "[ROUTER]: Recreating service instance %s with domain %s", id, domain.Domain)
	service, err := rs.serviceManager.RecreateWithDomain(id, domain)
	if err != nil {
		events.Logf(events.LOG_ERROR, "[ROUTER]: Error while recreating service instance for %s: %v", id, err)
		return err
	}

	config := rs.cloudMap[strings.ToLower(service.Name)]
	config.Domain = domain

	rs.cloudServiceInstanceMap[strings.ToLower(service.Name)] = service
	rs.cloudMap[strings.ToLower(service.Name)] = config

	events.Logf(events.LOG_INFO, "[ROUTER]: Service instance for %s recreated successfully", service.Name)

	err = rs.globalConfig.PersistConfig()
	if err != nil {
		events.Logf(events.LOG_ERROR, "[ROUTER]: Error while persisting router config: %v", err)
	}
	return nil
}

func (rs *RouterState) ListServiceInstances() []*cloud.ServiceInstance {
	return rs.serviceManager.ListInstances()
}

func (rs *RouterState) StartInstance(id string) error {
	events.Logf(events.LOG_INFO, "[ROUTER]: Starting service instance %s", id)
	return rs.serviceManager.StartInstance(id)
}

func (rs *RouterState) StopInstance(id string) error {
	events.Logf(events.LOG_INFO, "[ROUTER]: Stopping service instance %s", id)
	return rs.serviceManager.StopInstance(id)
}

func (rs *RouterState) RestartInstance(id string) error {
	events.Logf(events.LOG_INFO, "[ROUTER]: Restarting service instance %s", id)
	return rs.serviceManager.RestartInstance(id)
}

func (rs *RouterState) DeleteInstance(id string) error {
	events.Logf(events.LOG_INFO, "[ROUTER]: Deleting service instance %s", id)
	return rs.serviceManager.DeleteInstance(id)
}

func (rs *RouterState) GetInstanceLogs(id string, tail int) (string, error) {
	return rs.serviceManager.GetInstanceLogs(id, tail)
}

func (rs *RouterState) InspectInstance(id string) error {
	return rs.serviceManager.RefreshInstanceStatus(id)
}

func (rs *RouterState) StreamInstanceLogs(id string) (io.Reader, error) {
	return rs.serviceManager.StreamInstanceLogs(id)
}

// Config

func (rs *RouterState) GetConfig() *Config {
	return rs.globalConfig
}

func (cf *Config) BuildVars() {
	for k, v := range cf.Variables {
		err := config.SetEnv(k, v)
		if err != nil {
			events.Logf(events.LOG_ERROR, "[ROUTER]: Error while setting environment variable %s: %v", k, err)
		}
	}
}

func (cf *Config) PersistConfig() error {
	_, err := config.CreateConfigDir(config.BASE_CONFIG_DIR)
	if err != nil {
		return err
	}
	configBytes, err := yaml.Marshal(cf)
	if err != nil {
		return err
	}
	events.Logf(events.LOG_INFO, "[ROUTER]: Persisting router config")
	return config.WriteIntoConfigFile(ROUTER_CONFIG_FILE, configBytes)
}
