// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewServerPool(t *testing.T) {
	sp := NewServerPool()
	assert.NotNil(t, sp)
	assert.NotNil(t, sp.servers)
}

func TestServerPool_AddGetDelServer(t *testing.T) {
	sp := NewServerPool()
	server := &Server{Name: "test", Protocol: "http", Host: "localhost", Port: 8080}
	uid := sp.AddNewServer(server)
	assert.NotEmpty(t, uid)

	got, err := sp.GetServer(uid)
	assert.NoError(t, err)
	assert.Equal(t, server, got)

	sp.DelServer(uid)
	_, err = sp.GetServer(uid)
	assert.Error(t, err)
}

func TestServerPool_GetServerWithName(t *testing.T) {
	sp := NewServerPool()
	server := &Server{Name: "test", Protocol: "http", Host: "localhost", Port: 8080}
	uid := sp.AddNewServer(server)

	got, err := sp.GetServerWithName("test")
	assert.NoError(t, err)
	assert.Equal(t, server, got)

	uid2, err := sp.GetServerUIDWithName("test")
	assert.NoError(t, err)
	assert.Equal(t, uid, uid2)
}

func TestServerPool_GetAllServer(t *testing.T) {
	sp := NewServerPool()
	server1 := &Server{Name: "s1"}
	server2 := &Server{Name: "s2"}
	sp.AddNewServer(server1)
	sp.AddNewServer(server2)
	all := sp.GetAllServer()
	assert.Len(t, all, 2)
}

func TestRoundRobinBalancer_GetNextServer(t *testing.T) {
	sp := NewServerPool()
	server1 := &Server{Name: "s1"}
	server2 := &Server{Name: "s2"}
	sp.AddNewServer(server1)
	sp.AddNewServer(server2)
	rr := NewRoundRobinBalancer(sp)

	srv, err := rr.GetNextServer()
	assert.NoError(t, err)
	assert.NotNil(t, srv)
	srv2, err := rr.GetNextServer()
	assert.NoError(t, err)
	assert.NotEqual(t, srv, srv2)
}

func TestLoadBalancer_Serve(t *testing.T) {
	// Setup a dummy backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer backend.Close()

	server := &Server{Name: "backend", URL: backend.URL}
	sp := NewServerPool()
	sp.AddNewServer(server)
	rr := NewRoundRobinBalancer(sp)
	lb := NewLoadBalancer(rr)
	lb.Logs = make(chan Logs, 10)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	lb.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode)
}

func TestServerUpgradeProxy(t *testing.T) {
	sp := NewServerPool()
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	defer backend1.Close()
	defer backend2.Close()
	server1 := &Server{Name: "test1", URL: backend1.URL}
	sever2 := &Server{Name: "test2", URL: backend2.URL}
	backend1Uri, err := url.Parse(backend1.URL)

	sp.AddNewServer(server1)
	sp.AddNewServer(sever2)

	assert.Equal(t, nil, err)

	backend1Proxy := NewProxy(backend1Uri)
	server1.Proxy = backend1Proxy

	err = server1.UpgradeProxy()

	assert.Equal(t, nil, err)

	rr := NewRoundRobinBalancer(sp)
	lb := NewLoadBalancer(rr)
	lb.Logs = make(chan Logs, 10)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	lb.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPing(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/ping", nil)
	Ping(w, r)
	assert.Equal(t, "pong", w.Body.String())
}

func TestBuildServerURL(t *testing.T) {
	server := &Server{Protocol: "http", Host: "localhost", Port: 8080}
	url, err := buildServerURL(server)
	assert.NoError(t, err)
	assert.Contains(t, url, "http://localhost:8080")
}

func TestSerializeHealthCheckStatus(t *testing.T) {
	status := &HealthCheckStatus{
		Pass:     []ServerStatus{{Name: "s1", Url: "u1", Healthy: true}},
		Fail:     []ServerStatus{{Name: "s2", Url: "u2", Healthy: false}},
		Duration: time.Second,
	}
	s, err := SerializeHealthCheckStatus(status)
	assert.NoError(t, err)
	assert.Contains(t, s, "s1")
	assert.Contains(t, s, "s2")
}
