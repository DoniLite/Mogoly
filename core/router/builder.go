package router

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/DoniLite/Mogoly/cloud"
	"github.com/DoniLite/Mogoly/core/events"
	"github.com/DoniLite/Mogoly/core/server"
)

var (
	currentRouter *RouterState
)

func Startup(config *Config) {
	// First trying to load the config from the file
	configFromFile, err := LoadConfig()

	if configFromFile != nil && err == nil {
		buildRouter(configFromFile)
	}

	if config != nil {
		MergeRouterConfigs(config)
	}
}

func buildRouter(initialConfig *Config) {
	var err error
	manager, err := cloud.NewCloudDBManager()
	if err != nil {
		panic(err)
	}

	events.Logf(events.LOG_INFO, "[ROUTER]: Building new router for the global server state")

	rs := &RouterState{
		httpServerMap: make(map[string]http.Handler),
		serverMap:     make(map[string]*server.Server),
		cloudServiceInstanceMap: make(map[string]*cloud.ServiceInstance),
		cloudMap:                make(map[string]*cloud.ServiceConfig),
		serviceManager:          manager,
	}
	for _, server := range initialConfig.Servers {
		if _, exists := rs.serverMap[strings.ToLower(server.Name)]; exists {
			continue
		}
		rs.httpServerMap[strings.ToLower(server.Name)] = CreateSingleHttpServer(server)
		rs.serverMap[strings.ToLower(server.Name)] = server
	}

	for _, service := range initialConfig.Services {
		if _, exists := rs.cloudMap[strings.ToLower(service.Name)]; exists {
			continue
		}
		rs.cloudMap[strings.ToLower(service.Name)] = service
		currentService, err := rs.serviceManager.CreateInstance(*service)
		if err != nil {
			events.Logf(events.LOG_ERROR, "[ROUTER]: Error while creating service instance for %s: %v", service.Name, err)
			continue
		}
		rs.cloudServiceInstanceMap[strings.ToLower(service.Name)] = currentService
	}

	rs.globalConfig = initialConfig
	currentRouter = rs
	events.Logf(events.LOG_INFO, "[ROUTER]: New router builded and assigned correctly")
}

func GetRouter() (*RouterState, error) {
	if currentRouter == nil {
		return nil, fmt.Errorf("router not initialized")
	}
	return currentRouter, nil
}

func CreateSingleHttpServer(s *server.Server) http.Handler {
	events.Logf(events.LOG_INFO, "[SERVER]: Creating new http handler for %s server", s.Name)
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.ServeHTTP)

	var middlewares []func(http.Handler) http.Handler

	events.Logf(events.LOG_INFO, "[SERVER]: Assigning middlewares for the %s server", s.Name)
	for _, v := range s.Middlewares {
		m := server.MiddlewaresList[server.MiddleWareName(v.Name)]

		middlewares = append(middlewares, m.Fn(v.Config))
	}
	events.Logf(events.LOG_INFO, "[SERVER]: %d assigned to the %s server", len(middlewares), s.Name)

	return server.ChainMiddleware(mux, middlewares...)
}

func MergeRouterConfigs(newConfig *Config) *Config {
	conf := currentRouter.globalConfig

	if conf == nil {
		conf = &Config{}
	}

	for _, srv := range newConfig.Servers {
		if !slices.ContainsFunc(conf.Servers, func(s *server.Server) bool {
			return s.Name == srv.Name
		}) {
			conf.Servers = append(conf.Servers, srv)
			currentRouter.AddServer(srv)
		}
	}

	for _, svc := range newConfig.Services {
		if !slices.ContainsFunc(conf.Services, func(s *cloud.ServiceConfig) bool {
			return s.Name == svc.Name
		}) {
			conf.Services = append(conf.Services, svc)
		}
	}

	for key, value := range newConfig.Variables {
		if _, exists := conf.Variables[key]; !exists {
			conf.Variables[key] = value
		}
	}
	currentRouter.globalConfig = conf
	return conf
}
