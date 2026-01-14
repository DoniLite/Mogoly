// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoniLite/Mogoly/core/server"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_ConfigParsing(t *testing.T) {
	jsonContent := []byte(`{"server":[{"name":"s1","protocol":"http","host":"localhost","port":8080}]}`)
	cfg, err := ParseConfig(jsonContent, "json")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(cfg.Servers))

	badContent := []byte(`{invalid}`)
	_, err2 := ParseConfig(badContent, "json")
	assert.Error(t, err2)

	_, err3 := ParseConfig(jsonContent, "xml")
	assert.Error(t, err3)
}

func TestIntegration_HealthChecker_Error(t *testing.T) {
	s := &server.Server{Name: "bad", URL: "http://invalid:9999"}
	ok, err := server.HealthChecker(s)
	assert.False(t, ok)
	assert.Error(t, err)
}

func TestIntegration_LoadBalancerWithHealthCheck(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("healthy"))
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}
	}))
	defer healthy.Close()

	unhealthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer unhealthy.Close()

	s := &server.Server{Name: "main"}
	serverHealthy := &server.Server{Name: "healthy", URL: healthy.URL}
	serverUnhealthy := &server.Server{Name: "unhealthy", URL: unhealthy.URL}

	_ = serverHealthy.UpgradeProxy()
	_ = serverUnhealthy.UpgradeProxy()

	s.AddNewBalancingServer(serverHealthy)
	s.AddNewBalancingServer(serverUnhealthy)

	// Health check all
	status, err := s.CheckHealthAll()
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, len(status.Pass)+len(status.Fail) == 2)

	// Use only healthy server for load balancer
	s.BalancingServers = []*server.Server{serverHealthy}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode)
}
