package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"time"

	"github.com/DoniLite/Mogoly/core/config"
	"github.com/DoniLite/Mogoly/core/events"
)

// UpgradeProxy ensures server.Proxy is initialized for this server.
func (server *Server) UpgradeProxy() error {
	if server == nil {
		return errors.New("nil receiver: server")
	}
	server.logf(events.LOG_INFO, "[SERVER]: Checking the reversed proxy for the server %s", server.Name)
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.proxy != nil {
		server.logf(events.LOG_INFO, "[SERVER]: Existent proxy detected for the server %s", server.Name)
		return nil
	}
	serverURL, err := BuildServerURL(server)
	if err != nil {
		return err
	}
	u, err := url.Parse(serverURL)
	if err != nil {
		return err
	}
	server.logf(events.LOG_INFO, "[SERVER]: Upgrading the reversed proxy for the server %s", server.Name)
	server.proxy = NewProxy(u)
	return nil
}

func (server *Server) AddNewBalancingServer(bs *Server) {
	if bs == nil {
		server.logf(events.LOG_INFO, "[SERVER]: Empty balancing server received for the server %s", server.Name)
		return
	}
	server.logf(events.LOG_INFO, "[SERVER]: Adding new balancing server %s to %s", bs.Name, server.Name)
	server.BalancingServers = append(server.BalancingServers, bs)
}

func (server *Server) DelBalancingServer(name string) {
	if name == "" {
		return
	}
	filtered := server.BalancingServers[:0]
	for _, s := range server.BalancingServers {
		if s != nil && s.Name != name {
			filtered = append(filtered, s)
		}
	}
	server.BalancingServers = filtered
	server.logf(events.LOG_INFO, "[SERVER]: Removed server %s", name)
}

func (server *Server) GetServer(name string) *Server {
	if name == "" {
		return nil
	}
	for _, s := range server.BalancingServers {
		if s != nil && s.Name == name {
			return s
		}
	}
	return nil
}

// CheckHealthAll performs health checks without holding the lock during network I/O.
func (server *Server) CheckHealthAll() (*HealthCheckStatus, error) {
	server.mu.Lock()
	servers := slices.Clone(server.BalancingServers)
	server.mu.Unlock()

	server.logf(events.LOG_INFO, "[SERVER]: Preparing health checking for the %s server instances", server.Name)
	var hc HealthCheckStatus
	start := time.Now()

	for _, target := range servers {
		if target == nil {
			continue
		}
		checkStart := time.Now()
		success, err := HealthChecker(target)
		u, _ := BuildServerURL(target)

		target.mu.Lock()
		target.IsHealthy = (err == nil && success)
		t := checkStart
		target.LastHealthCheck = &t
		target.mu.Unlock()

		entry := ServerStatus{Name: target.Name, Url: u, Healthy: target.IsHealthy}
		if target.IsHealthy {
			hc.Pass = append(hc.Pass, entry)
		} else {
			hc.Fail = append(hc.Fail, entry)
		}
	}

	hc.Duration = time.Since(start)
	server.logf(events.LOG_INFO, "[SERVER]: Finished health check for the %s server instances in %vs", server.Name, hc.Duration.Seconds())
	return &hc, nil
}

func (server *Server) CheckHealthAny(name string) (*ServerStatus, error) {
	if name == "" {
		return nil, fmt.Errorf("empty server name")
	}
	server.mu.Lock()
	target := server.GetServer(name)
	server.mu.Unlock()
	if target == nil {
		return nil, fmt.Errorf("no server found for name %q", name)
	}
	u, _ := BuildServerURL(target)
	success, err := HealthChecker(target)
	healthy := err == nil && success
	return &ServerStatus{Name: target.Name, Url: u, Healthy: healthy}, err
}

func (server *Server) CheckHealthSelf() (*ServerStatus, error) {
	u, _ := BuildServerURL(server)
	success, err := HealthChecker(server)
	healthy := err == nil && success
	return &ServerStatus{Name: server.Name, Url: u, Healthy: healthy}, err
}

// RollBack replaces the current balancing set with the provided list atomically.
func (server *Server) RollBack(servers []*Server) {
	server.mu.Lock()
	server.logf(events.LOG_INFO, "[PROXY]: Rollback for the %s server with %v", server.Name, servers)
	server.BalancingServers = slices.Clone(servers)
	server.mu.Unlock()
}

// RollBackAny replaces a named server with a new one, or appends when name is empty.
func (server *Server) RollBackAny(name string, newServer *Server) error {
	if newServer == nil {
		return fmt.Errorf("newServer must not be nil")
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	if name == "" {
		server.AddNewBalancingServer(newServer)
		return nil
	}
	for i, s := range server.BalancingServers {
		if s != nil && s.Name == name {
			server.logf(events.LOG_INFO, "[PROXY]: Rollback %s server with %v", name, newServer)
			server.BalancingServers[i] = newServer
			return nil
		}
	}
	return fmt.Errorf("server named %q not found", name)
}

// GetNextServer returns the next healthy server using round-robin.
// If all are unhealthy or list is empty, an error is returned.
func (server *Server) GetNextServer(strategy ServerStrategy) (*Server, error) {
	switch strategy {
	case RoundRobin:
		return RoundRobinStrategy(server)
	case LeastConnections:
		return LeastConnectionsStrategy(server)
	case Random:
		return RandomStrategy(server)
	default:
		return nil, fmt.Errorf("unknown strategy: %s", strategy)
	}
}

// ServeHTTP implements a minimal LB+proxy. It avoids mutating the receiver on retry
// and constructs the target URL using ResolveReference to handle paths and queries correctly.
func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		backend *Server
		err     error
	)

	server.logf(events.LOG_INFO, "[Proxy]: New incoming request <- %s Method for %s: %s", r.URL.Path, r.Method, server.Name)
	server.logf(events.LOG_INFO, "[Load Balancer]: Selecting next server for the %s proxy", server.Name)

	if len(server.BalancingServers) > 0 {
		strategy := config.GetEnv(config.BALANCER_STRATEGY, string(RoundRobin))
		backend, err = server.GetNextServer(ServerStrategy(strategy))
		if err != nil {
			server.logf(events.LOG_ERROR, "[Load Balancer] error for the %s server: %v", server.Name, err)
			http.Error(w, "No backend available", http.StatusServiceUnavailable)
			return
		}
		if upErr := backend.UpgradeProxy(); upErr != nil {
			server.logf(events.LOG_ERROR, "failed to init backend proxy for the %s server: %v", backend.Name, upErr)
			http.Error(w, "No proxy service available", http.StatusInternalServerError)
			return
		}
	} else {
		// Single-node mode: proxy to self
		backend = server
		if upErr := server.UpgradeProxy(); upErr != nil {
			server.logf(events.LOG_ERROR, "failed to init proxy for the %s server: %v", server.Name, upErr)
			http.Error(w, "No proxy service available", http.StatusInternalServerError)
			return
		}
	}

	baseURL, err := parseServerURL(backend)
	if err != nil {
		server.logf(events.LOG_ERROR, "[Proxy]: invalid backend URL for %s: %v", backend.Name, err)
		http.Error(w, "Invalid backend url", http.StatusInternalServerError)
		return
	}

	// Build destination URL preserving path and query
	joinedPath := singleSlashJoin(baseURL.Path, r.URL.Path)
	target := *baseURL
	target.Path = joinedPath
	target.RawQuery = r.URL.RawQuery

	// Clone request with context and body; copy headers
	req, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), r.Body)
	if err != nil {
		server.logf(events.LOG_ERROR, "[Fatal]: cannot create outbound request for %s server: %v", server.Name, err)
		http.Error(w, "Failed to create backend request", http.StatusInternalServerError)
		return
	}
	req.Header = r.Header.Clone()
	appendForwardHeaders(req.Header, r, baseURL.Scheme)

	server.logf(events.LOG_INFO, "[Proxy]: Forwarding %s -> %s (backend Name: %s)", r.URL.String(), target.String(), backend.Name)

	// Delegate to the preconfigured reverse proxy for the backend.
	backend.proxy.ServeHTTP(w, req)
}

func (s *Server) logf(level events.LogType, format string, args ...any) {
	if s == nil {
		return
	}
	// Non-blocking send; drop if channel is full
	events.Logf(level, format, args...)
}
