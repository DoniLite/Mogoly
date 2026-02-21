package daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DoniLite/Mogoly/cli/actions"
)

func TestDaemonPing(t *testing.T) {
	// Setup temporary socket path
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "mogoly_test.sock")
	defer os.Remove(socketPath)

	// Create and start server
	server, err := NewServer(socketPath)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	go func() {
		if err := server.Start(); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()
	defer server.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Create client
	client := NewClient(socketPath)
	// No defer client.Close() because Close() currently just closes syncClient which we might want to do explicitly
	defer client.Close()

	// Test Ping
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

func TestDaemonStatus(t *testing.T) {
	// Setup temporary socket path
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "mogoly_test_status.sock")
	defer os.Remove(socketPath)

	// Create and start server
	server, err := NewServer(socketPath)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	go func() {
		if err := server.Start(); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()
	defer server.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Create client
	client := NewClient(socketPath)
	defer client.Close()

	// Test Status
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.SendAction(ctx, actions.ActionDaemonStatus, nil)
	if err != nil {
		t.Fatalf("SendAction failed: %v", err)
	}

	if resp.Error != "" {
		t.Fatalf("Daemon returned error: %s", resp.Error)
	}

	var status map[string]interface{}
	if err := DecodePayload(resp, &status); err != nil {
		t.Fatalf("Failed to decode payload: %v", err)
	}

	if status["socket"] != socketPath {
		t.Errorf("Expected socket path %s, got %v", socketPath, status["socket"])
	}
}
