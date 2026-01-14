package router

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DoniLite/Mogoly/core/domain"
	"github.com/DoniLite/Mogoly/core/server"
)

func TestServeHTTPS_LocalhostSNI(t *testing.T) {
	// Setup temp home for config
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// Setup temp hosts file
	hostsPath := filepath.Join(tempHome, "hosts")
	if err := os.WriteFile(hostsPath, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create hosts file: %v", err)
	}
	t.Setenv("MOGOLY_HOSTS_PATH", hostsPath)

	// backend server "app.localhost" served over Mogoly HTTPS
	app := httptest.NewServer(http.HandlerFunc(server.Ping))
	defer app.Close()
	s := &server.Server{Name: "app.localhost", URL: app.URL}
	BuildRouter(&Config{Servers: []*server.Server{s}})

	cm, err := domain.NewManager()
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	if err := cm.Add("app.localhost", true); err != nil {
		t.Fatalf("failed to add domain: %v", err)
	}
	ts := ServeHTTPS("127.0.0.1:0", cm)
	defer func() {
		if err := ts.Close(); err != nil {
			log.Printf("Error while closing the server: %v", err)
		}
	}()

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
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Error while closing the body reader: %v", err)
		}
	}()
	if res.StatusCode != 200 {
		t.Fatalf("want 200, got %d", res.StatusCode)
	}
}
