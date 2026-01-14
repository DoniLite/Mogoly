package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/DoniLite/Mogoly/core/events"
	"github.com/DoniLite/Mogoly/core/server"
)

var (
	currentRouter *RouterState
)

func BuildRouter(config *Config) {
	events.Logf(events.LOG_INFO, "[ROUTER]: Building new router for the global server state")
	rs := &RouterState{
		m: make(map[string]http.Handler),
		s: make(map[string]*server.Server),
	}
	for _, server := range config.Servers {
		rs.m[strings.ToLower(server.Name)] = CreateSingleHttpServer(server)
		rs.s[strings.ToLower(server.Name)] = server
	}

	rs.globalConfig = config
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
