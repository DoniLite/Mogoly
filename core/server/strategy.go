package server

import (
	"fmt"

	"github.com/DoniLite/Mogoly/core/events"
)

const (
	RoundRobin       ServerStrategy = "round_robin"
	LeastConnections ServerStrategy = "least_connections"
	Random           ServerStrategy = "random"
)

func RoundRobinStrategy(server *Server) (*Server, error) {
	server.mu.Lock()
	defer server.mu.Unlock()

	n := len(server.BalancingServers)
	if n == 0 {
		return nil, fmt.Errorf("no backend servers configured")
	}
	// Iterate at most n times to find a healthy server.
	for range n {
		server.idx = (server.idx + 1) % n
		cand := server.BalancingServers[server.idx]
		if cand != nil && cand.IsHealthy { // prefer healthy
			server.logf(events.LOG_INFO, "[PROXY]: Next server found for the %s proxy with id %d", server.Name, server.idx)
			return cand, nil
		}
	}
	// Fallback: return next even if unhealthy to avoid total outage
	cand := server.BalancingServers[server.idx]
	if cand != nil {
		server.logf(events.LOG_INFO, "[PROXY]: Next unhealthy server found for the %s proxy with id %d", server.Name, server.idx)
		return cand, nil
	}
	return nil, fmt.Errorf("no usable backend server found")
}

func LeastConnectionsStrategy(server *Server) (*Server, error) {
	return nil, fmt.Errorf("not implemented")
}

func RandomStrategy(server *Server) (*Server, error) {
	return nil, fmt.Errorf("not implemented")
}

