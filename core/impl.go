// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

// Server pooling implementation

func (server *Server) UpgradeProxy() error {
	if server.Proxy != nil {
		return nil
	}
	serverUrl, err := buildServerURL(server)
	if err != nil {
		return err
	}
	proxyUrl, err := url.Parse(serverUrl)
	if err != nil {
		return err
	}
	proxy := NewProxy(proxyUrl)

	server.Proxy = proxy

	return nil
}

func (server *Server) AddNewBalancingServer(bs *Server) {
	server.BalancingServers = append(server.BalancingServers, bs)
}

func (server *Server) DelBalancingServer(name string) {
	var filteredServers []*Server

	for _, s := range server.BalancingServers {
		if s.Name != name {
			filteredServers = append(filteredServers, s)
		}
		continue
	}

	server.BalancingServers = filteredServers
}

func (server *Server) GetServer(name string) *Server {
	var targetServer *Server

	for _, s := range server.BalancingServers {
		if s.Name == name {
			targetServer = s
		}
		continue
	}

	return targetServer
}

func (server *Server) CheckHealthAll() (*HealthCheckStatus, error) {
	server.mu.Lock()
	defer server.mu.Unlock()
	var healthChecker HealthCheckStatus
	healthCheckStartTime := time.Now()

	for uid, server := range server.BalancingServers {
		serverHealCheckTime := time.Now()
		success, err := HealthChecker(server)
		url, _ := buildServerURL(server)

		if err != nil || !success {
			server.BalancingServers[uid].IsHealthy = false
			server.BalancingServers[uid].LastHealthCheck = &serverHealCheckTime

			healthChecker.Fail = append(healthChecker.Fail, ServerStatus{
				Name:    server.Name,
				Url:     url,
				Healthy: false,
			})
		} else {
			server.BalancingServers[uid].IsHealthy = true
			server.BalancingServers[uid].LastHealthCheck = &serverHealCheckTime

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

func (server *Server) CheckHealthAny(name string) (*ServerStatus, error) {
	server.mu.Lock()
	defer server.mu.Unlock()

	if targetServer := server.GetServer(name); targetServer != nil {
		url, _ := buildServerURL(targetServer)
		if success, err := HealthChecker(targetServer); !success || err != nil {
			return &ServerStatus{
				Name:    targetServer.Name,
				Url:     url,
				Healthy: false,
			}, err
		}

		return &ServerStatus{
			Name:    targetServer.Name,
			Url:     url,
			Healthy: true,
		}, nil
	}

	return nil, fmt.Errorf("no server found for the %s name", name)
}

func (server *Server) CheckHealthSelf() (*ServerStatus, error) {
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

func (server *Server) RollBack(servers []*Server) {
	server.mu.Lock()
	defer server.mu.Unlock()
	previousServer := slices.Clone(server.BalancingServers)

	for _, value := range servers {
		server.AddNewBalancingServer(value)
	}

	for _, s := range previousServer {
		server.DelBalancingServer(s.Name)
	}
}

func (server *Server) RollBackAny(name string, newServer *Server) error {
	server.mu.Lock()
	defer server.mu.Unlock()

	if name == "" && newServer != nil {
		server.AddNewBalancingServer(newServer)
		return nil
	}

	if name != "" && newServer != nil {
		server.DelBalancingServer(name)
		server.AddNewBalancingServer(newServer)
		return nil
	}

	if name != "" && newServer == nil {
		return fmt.Errorf("provide a name %s but don't provide any server for the roll backing", name)
	}

	return fmt.Errorf("invalid argument provided for name %s and server %v", name, newServer)

}

// Round Robin Balancer implementation

func (server *Server) GetNextServer() (*Server, error) {

	server.mu.Lock()
	defer server.mu.Unlock()

	server.idx = (server.idx + 1) % len(server.BalancingServers)

	selectedServer := server.BalancingServers[server.idx]

	return selectedServer, nil
}

//  Load Balancer

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var balancingServer *Server
	var err error
	var targetURL *url.URL
	var req *http.Request
	var targetPath string

	server.Logs <- Logs{
		message: fmt.Sprintf("[Proxy]: New incoming request <- %s Method: %s", r.URL.Path, r.Method),
		logType: LOG_INFO,
	}

	server.Logs <- Logs{
		message: "[Load Balancer]: Trying to get the next server",
		logType: LOG_ERROR,
	}

	if len(server.BalancingServers) > 1 {
		balancingServer, err = server.GetNextServer()

		if err != nil {
			server.Logs <- Logs{
				message: "[Load Balancer]: Failed to get the next server new attempt",
				logType: LOG_ERROR,
			}
			// Trying to get the next server after this
			server, err = server.GetNextServer()
			if err != nil {
				server.Logs <- Logs{
					message: "[Load Balancer]: Failed to get the next server (second attempt)",
					logType: LOG_ERROR,
				}
				http.Error(w, "Internal server error", 500)
				return
			}
		}

		err = balancingServer.UpgradeProxy()

		if err != nil {
			server.Logs <- Logs{
				message: "Failed to upgrade the server proxy",
				logType: LOG_ERROR,
			}
		}
	}

	err = server.UpgradeProxy()

	if err != nil {
		server.Logs <- Logs{
			message: "Failed to upgrade the server proxy",
			logType: LOG_ERROR,
		}
	}

	if balancingServer != nil {
		targetURL, err = url.Parse(balancingServer.URL)

		if server.URL == "" || err != nil {

			targetURL, err = url.Parse(fmt.Sprintf("%s://%s:%d", balancingServer.Protocol, balancingServer.Host, balancingServer.Port))

			if err != nil {
				server.Logs <- Logs{
					message: fmt.Sprintf("[Proxy]: Invalid url: %s from the incoming request [Host] %s", targetURL, r.Host),
					logType: LOG_ERROR,
				}
				http.Error(w, "Invalid backend url", http.StatusInternalServerError)
				return
			}
		}

		if err != nil {
			server.Logs <- Logs{
				message: fmt.Sprintf("[Proxy]: Invalid url: %s from the incoming request [Host] %s", targetURL, r.Host),
				logType: LOG_ERROR,
			}
			http.Error(w, "Invalid backend url", http.StatusInternalServerError)
			return
		}

		targetPath = strings.TrimPrefix(targetURL.String(), "/") + r.URL.Path
	} else {

		targetURL, err = url.Parse(server.URL)

		if server.URL == "" || err != nil {

			targetURL, err = url.Parse(fmt.Sprintf("%s://%s:%d", server.Protocol, server.Host, server.Port))

			if err != nil {
				server.Logs <- Logs{
					message: fmt.Sprintf("[Proxy]: Invalid url: %s from the incoming request [Host] %s", targetURL, r.Host),
					logType: LOG_ERROR,
				}
				http.Error(w, "Invalid backend url", http.StatusInternalServerError)
				return
			}
		}

		if err != nil {
			server.Logs <- Logs{
				message: fmt.Sprintf("[Proxy]: Invalid url: %s from the incoming request [Host] %s", targetURL, r.Host),
				logType: LOG_ERROR,
			}
			http.Error(w, "Invalid backend url", http.StatusInternalServerError)
			return
		}

		targetPath = strings.TrimPrefix(targetURL.String(), "/") + r.URL.Path
	}

	server.Logs <- Logs{
		message: fmt.Sprintf("[Proxy]: Preparing new request for the URI: %s", targetPath),
		logType: LOG_INFO,
	}

	if r.URL.RawQuery != "" {
		targetPath += "?" + r.URL.RawQuery
	}

	req, err = http.NewRequest(r.Method, targetPath, r.Body)

	if err != nil {
		server.Logs <- Logs{
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

	req.Header.Set("X-Forwarded-For", r.RemoteAddr)

	server.Logs <- Logs{
		message: fmt.Sprintf("[Proxy]: Forwarding new request <-> %s to the backend server ID: %s", req.URL.String(), server.ID),
		logType: LOG_INFO,
	}

	balancingServer.Proxy.ServeHTTP(w, req)
}
