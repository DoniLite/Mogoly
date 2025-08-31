package core

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxy_RoundRobin_Forwarding(t *testing.T) {
	b1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "one") }))
	defer b1.Close()
	b2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "two") }))
	defer b2.Close()
	var got string
	var firstHitsGot string

	a := &Server{Name: "a", URL: b1.URL, IsHealthy: true}
	b := &Server{Name: "b", URL: b2.URL, IsHealthy: true}
	lb := &Server{Name: "lb", BalancingServers: []*Server{a, b}}

	// handler under test
	h := http.HandlerFunc(lb.ServeHTTP)

	// 1st hits a
	req := httptest.NewRequest("GET", "http://x/internal", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	got = rr.Body.String()
	firstHitsGot = got

	// 2nd hits b
	req = httptest.NewRequest("GET", "http://x/internal", nil)
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if got = rr.Body.String(); got == firstHitsGot {
		t.Fatalf("2nd -> want other than gotten, got %q", got)
	}

	// mark b unhealthy => next should return a again
	b.IsHealthy = false
	req = httptest.NewRequest("GET", "http://x/internal", nil)
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if got := rr.Body.String(); got != "one" {
		t.Fatalf("3rd -> want one, got %q", got)
	}
}
