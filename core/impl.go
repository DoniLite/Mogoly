package core

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// Server pooling implementation

func (sp *ServerPool) AddNewServer(server *Server) string {
	serverUUID := uuid.NewString()
	sp.servers[serverUUID] = server
	return serverUUID
}

func (sp *ServerPool) GetServer(uuid string) (*Server, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if server, ok := sp.servers[uuid]; ok {
		return server, nil
	}
	return nil, fmt.Errorf("this server %s not exist", uuid)
}

func (sp *ServerPool) GetAllServer() []*Server {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	var servers []*Server
	for _, s := range sp.servers {
		servers = append(servers, s)
	}
	return servers
}

// Round Robin Balancer implementation

func (rb *RoundRobinBalancer) GetNextServer() (*Server, error) {
	servers := rb.pool.GetAllServer()

	if len(servers) == 0 {
		return nil, fmt.Errorf("no server found try to add server to your Server pool before")
	}

	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.idx = (rb.idx + 1) % len(servers)

	selectedServer := servers[rb.idx]

	return selectedServer, nil
}

//  Load Balancer

func (lb *LoadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
	var targetURL *url.URL
	server, err := lb.strategy.GetNextServer()

	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}

	targetURL, err = url.Parse(server.URL)

	if server.URL == "" {
		
		targetURL, err = url.Parse(fmt.Sprintf("%s://%s:%d", server.Protocol, server.Host, server.Port))

		if err != nil {
			http.Error(w, "Invalid backend url", http.StatusInternalServerError)
			return
		}
	}

	if err != nil {
		http.Error(w, "Invalid backend url", http.StatusInternalServerError)
		return
	}

	targetPath := strings.TrimPrefix(targetURL.String(), "/") + r.URL.Path

	if r.URL.RawQuery != "" {
		targetPath += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequest(r.Method, targetPath, r.Body)

	if err != nil {
		http.Error(w, "Failed to create the request to the backend server", http.StatusInternalServerError)
		return
	}

	for k, values := range r.Header {
		for _, value := range values {
			req.Header.Add(k, value)
		}
	}

	req.Header.Set("X-Forwarded-for", r.RemoteAddr)

	lb.proxy.ServeHTTP(w, req)
}
