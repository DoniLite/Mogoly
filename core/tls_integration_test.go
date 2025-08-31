package core

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServeHTTPS_LocalhostSNI(t *testing.T) {
	// backend server "app.localhost" served over Mogoly HTTPS
	app := httptest.NewServer(http.HandlerFunc(Ping))
	defer app.Close()
	s := &Server{Name: "app.localhost", URL: app.URL}
	BuildRouter(&Config{Servers: []*Server{s}})

	cm := NewCertManager(t.TempDir(), "noreply@example.com", "development")
	ts := ServeHTTPS("127.0.0.1:0", cm)
	defer ts.Close()

	// build client that trusts self-signed (skip verify for test)
	addr := ts.Addr
	url := fmt.Sprintf("https://%s/", addr)
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true, ServerName: "app.localhost"}}
	client := &http.Client{Transport: tr, Timeout: 3 * time.Second}

	req, _ := http.NewRequest("GET", url, nil)
	req.Host = "app.localhost"

	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("https request: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("want 200, got %d", res.StatusCode)
	}
}
