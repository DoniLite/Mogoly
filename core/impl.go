package core

import (
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Server pooling implementation

func (sp *ServerPool) AddNewServer(server *Server) string {
	serverUUID := uuid.NewString()
	serverId := len(sp.servers) - 1
	server.ID = fmt.Sprint(serverId)
	sp.servers[serverUUID] = server
	return serverUUID
}

func (sp *ServerPool) DelServer(uid string) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	delete(sp.servers, uid)
}

func (sp *ServerPool) GetServer(uuid string) (*Server, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if server, ok := sp.servers[uuid]; ok {
		return server, nil
	}
	return nil, fmt.Errorf("this server %s not exist", uuid)
}

func (sp *ServerPool) GetServerWithName(name string) (*Server, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	var foundedServer *Server

	for _, server := range sp.servers {
		if server.Name == name {
			foundedServer = server
			break
		}
		continue
	}

	if foundedServer == nil {
		return nil, fmt.Errorf("Server with name %s doesn't exist", name)
	}

	return foundedServer, nil
}

func (sp *ServerPool) GetServerUIDWithName(name string) (string, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	var foundedUID string

	for uid, server := range sp.servers {
		if server.Name == name {
			foundedUID = uid
			break
		}
		continue
	}

	if foundedUID == "" {
		return "", fmt.Errorf("Server with name %s doesn't exist", name)
	}

	return foundedUID, nil
}

func (sp *ServerPool) CheckHealthAll() (*HealthCheckStatus, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	var healthChecker HealthCheckStatus
	healthCheckStartTime := time.Now()

	for uid, server := range sp.servers {
		serverHealCheckTime := time.Now()
		success, err := HealthChecker(server)
		url, _ := buildServerURL(server)

		if err != nil || !success {
			sp.servers[uid].IsHealthy = false
			sp.servers[uid].LastHealthCheck = serverHealCheckTime

			healthChecker.Fail = append(healthChecker.Fail, ServerStatus{
				Name:    server.Name,
				Url:     url,
				Healthy: false,
			})
		} else {
			sp.servers[uid].IsHealthy = true
			sp.servers[uid].LastHealthCheck = serverHealCheckTime

			healthChecker.Pass = append(healthChecker.Pass, ServerStatus{
				Name:    server.Name,
				Url:     url,
				Healthy: true,
			})
		}
	}

	duration := time.Since(healthCheckStartTime)
	healthChecker.Duration = duration

	return &healthChecker, nil
}

func (sp *ServerPool) CheckHealthAny(uid string) (*ServerStatus, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if server, ok := sp.servers[uid]; ok {
		url, _ := buildServerURL(server)
		if success, err := HealthChecker(server); !success || err != nil {
			return &ServerStatus{
				Name:    server.Name,
				Url:     url,
				Healthy: false,
			}, err
		}

		return &ServerStatus{
			Name:    server.Name,
			Url:     url,
			Healthy: true,
		}, nil
	}

	return nil, fmt.Errorf("no server found for the %s uid", uid)
}

func (sp *ServerPool) RollBack(servers []*Server) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	previousServer := maps.Clone(sp.servers)

	for _, value := range servers {
		sp.AddNewServer(value)
	}

	for k := range previousServer {
		sp.DelServer(k)
	}
}

func (sp *ServerPool) RollBackAny(uid string, server *Server) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if uid == "" && server != nil {
		sp.AddNewServer(server)
		return nil
	}

	if uid != "" && server != nil {
		sp.DelServer(uid)
		sp.AddNewServer(server)
		return nil
	}

	if uid != "" && server == nil {
		return fmt.Errorf("provide an uid %s but don't provide any server for the roll backing", uid)
	}

	return fmt.Errorf("invalid argument provided for uid %s and server %v", uid, server)

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
	lb.Logs <- Logs{
		message: fmt.Sprintf("[Proxy]: New incoming request <- %s Method: %s", r.URL.Path, r.Method),
		logType: LOG_INFO,
	}
	var targetURL *url.URL
	lb.Logs <- Logs{
		message: "[Load Balancer]: Trying to get the next server",
		logType: LOG_ERROR,
	}
	server, err := lb.strategy.GetNextServer()

	if err != nil {
		lb.Logs <- Logs{
			message: "[Load Balancer]: Failed to get the next server new attempt",
			logType: LOG_ERROR,
		}
		// Trying to get the next server after this
		server, err = lb.strategy.GetNextServer()
		if err != nil {
			lb.Logs <- Logs{
				message: "[Load Balancer]: Failed to get the next server (second attempt)",
				logType: LOG_ERROR,
			}
			http.Error(w, "Internal server error", 500)
			return
		}
	}

	targetURL, err = url.Parse(server.URL)

	if server.URL == "" || err != nil {

		targetURL, err = url.Parse(fmt.Sprintf("%s://%s:%d", server.Protocol, server.Host, server.Port))

		if err != nil {
			lb.Logs <- Logs{
				message: fmt.Sprintf("[Proxy]: Invalid url: %s from the incoming request [Host] %s", targetURL, r.Host),
				logType: LOG_ERROR,
			}
			http.Error(w, "Invalid backend url", http.StatusInternalServerError)
			return
		}
	}

	if err != nil {
		lb.Logs <- Logs{
			message: fmt.Sprintf("[Proxy]: Invalid url: %s from the incoming request [Host] %s", targetURL, r.Host),
			logType: LOG_ERROR,
		}
		http.Error(w, "Invalid backend url", http.StatusInternalServerError)
		return
	}

	targetPath := strings.TrimPrefix(targetURL.String(), "/") + r.URL.Path

	lb.Logs <- Logs{
		message: fmt.Sprintf("[Proxy]: Preparing new request for the URI: %s", targetPath),
		logType: LOG_INFO,
	}

	if r.URL.RawQuery != "" {
		targetPath += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequest(r.Method, targetPath, r.Body)

	if err != nil {
		lb.Logs <- Logs{
			message: "[Fatal]: New forwarded request creation failed",
			logType: LOG_ERROR,
		}
		http.Error(w, "Failed to create the request to the backend server", http.StatusInternalServerError)
		return
	}

	for k, values := range r.Header {
		for _, value := range values {
			req.Header.Add(k, value)
		}
	}

	req.Header.Set("X-Forwarded-for", r.RemoteAddr)

	lb.Logs <- Logs{
		message: fmt.Sprintf("[Proxy]: Forwarding new request <-> %s to the backend server ID: %s", req.URL.String(), server.ID),
		logType: LOG_INFO,
	}

	lb.proxy.ServeHTTP(w, req)
}
