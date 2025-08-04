// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestIntegration_LoadBalancerWithHealthCheck(t *testing.T) {
	// Start two backend servers, one healthy, one unhealthy
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("healthy"))
	}))
	defer healthy.Close()

	unhealthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer unhealthy.Close()

	uHealthy, _ := url.Parse(healthy.URL)

	serverHealthy := &Server{Name: "healthy", URL: healthy.URL}
	serverUnhealthy := &Server{Name: "unhealthy", URL: unhealthy.URL}

	sp := NewServerPool()
	sp.AddNewServer(serverHealthy)
	sp.AddNewServer(serverUnhealthy)

	// Health check all
	status, err := sp.CheckHealthAll()
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, len(status.Pass)+len(status.Fail) == 2)

	// Use only healthy server for load balancer
	rr := NewRoundRobinBalancer(sp)
	proxy := NewProxy(uHealthy)
	lb := NewLoadBalancer(rr, proxy)
	lb.Logs = make(chan Logs, 10)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	lb.Serve(w, req)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode)
}

func TestIntegration_ConfigParsing(t *testing.T) {
	// Test config parsing for both JSON and YAML
	jsonContent := []byte(`{"proxy":{"name":"p1","host":"localhost","listen_port":"8080"},"server":[{"name":"s1","protocol":"http","host":"localhost","port":8080}]}`)
	cfg, err := ParseConfig(jsonContent, "json")
	assert.NoError(t, err)
	assert.Equal(t, "p1", cfg.Proxy.Name)
	assert.Equal(t, 1, len(cfg.Servers))
}
