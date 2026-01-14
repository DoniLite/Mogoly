// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package hosts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if m == nil {
		t.Fatal("Manager is nil")
	}

	if m.entries == nil {
		t.Error("Entries map not initialized")
	}

	if m.hostsPath == "" {
		t.Error("Hosts path is empty")
	}
}

func TestGetHostsFilePath(t *testing.T) {
	path := getHostsFilePath()
	if path == "" {
		t.Error("Hosts file path is empty")
	}
	t.Logf("Hosts file path: %s", path)
}

func TestGetLineEnding(t *testing.T) {
	ending := getLineEnding()
	if ending != "\n" && ending != "\r\n" {
		t.Errorf("Unexpected line ending: %q", ending)
	}
}

func TestManagerWithTempFile(t *testing.T) {
	// Create a temporary hosts file
	tempDir := t.TempDir()
	tempHostsFile := filepath.Join(tempDir, "hosts")

	initialContent := `127.0.0.1 localhost
::1 localhost
127.0.0.1 existing.local # Existing entry
`
	if err := os.WriteFile(tempHostsFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create temp hosts file: %v", err)
	}

	// Create manager with custom path
	m := &Manager{
		hostsPath: tempHostsFile,
		entries:   make(map[string]string),
	}

	if err := m.loadEntries(); err != nil {
		t.Fatalf("Failed to load entries: %v", err)
	}

	// Test adding an entry
	domain := "test.local"
	ip := "127.0.0.1"

	err := m.Add(domain, ip)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Verify entry was added
	if !m.Has(domain) {
		t.Error("Entry not found after adding")
	}

	storedIP, exists := m.Get(domain)
	if !exists {
		t.Error("Entry doesn't exist")
	}

	if storedIP != ip {
		t.Errorf("Stored IP = %s, want %s", storedIP, ip)
	}

	// Verify file was updated
	content, err := os.ReadFile(tempHostsFile)
	if err != nil {
		t.Fatalf("Failed to read hosts file: %v", err)
	}

	if !strings.Contains(string(content), domain) {
		t.Error("Domain not found in hosts file")
	}

	if !strings.Contains(string(content), MogolyMarker) {
		t.Error("Mogoly marker not found in hosts file")
	}

	// Test removing entry
	err = m.Remove(domain)
	if err != nil {
		t.Fatalf("Failed to remove entry: %v", err)
	}

	if m.Has(domain) {
		t.Error("Entry still exists after removal")
	}

	// Verify file was updated
	content, err = os.ReadFile(tempHostsFile)
	if err != nil {
		t.Fatalf("Failed to read hosts file: %v", err)
	}

	if strings.Contains(string(content), domain+" ") {
		t.Error("Domain still in hosts file after removal")
	}
}

func TestEmptyDomainValidation(t *testing.T) {
	tempDir := t.TempDir()
	tempHostsFile := filepath.Join(tempDir, "hosts")

	if err := os.WriteFile(tempHostsFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create temp hosts file: %v", err)
	}

	m := &Manager{
		hostsPath: tempHostsFile,
		entries:   make(map[string]string),
	}

	// Empty domain should fail
	err := m.Add("", "127.0.0.1")
	if err == nil {
		t.Error("Expected error when adding empty domain")
	}

	// Empty IP should fail
	err = m.Add("test.local", "")
	if err == nil {
		t.Error("Expected error when adding empty IP")
	}
}

func TestListEntries(t *testing.T) {
	tempDir := t.TempDir()
	tempHostsFile := filepath.Join(tempDir, "hosts")

	initialContent := `127.0.0.1 localhost
127.0.0.1 app1.local # Managed by Mogoly
127.0.0.1 app2.local # Managed by Mogoly
192.168.1.1 external.com # Not managed by Mogoly
`
	if err := os.WriteFile(tempHostsFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create temp hosts file: %v", err)
	}

	m := &Manager{
		hostsPath: tempHostsFile,
		entries:   make(map[string]string),
	}

	if err := m.loadEntries(); err != nil {
		t.Fatalf("Failed to load entries: %v", err)
	}

	entries := m.List()

	// Should only have Mogoly-managed entries
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	if _, exists := entries["app1.local"]; !exists {
		t.Error("app1.local not in entries")
	}

	if _, exists := entries["app2.local"]; !exists {
		t.Error("app2.local not in entries")
	}

	if _, exists := entries["external.com"]; exists {
		t.Error("external.com should not be in Mogoly entries")
	}
}

func TestConcurrency(t *testing.T) {
	tempDir := t.TempDir()
	tempHostsFile := filepath.Join(tempDir, "hosts")

	if err := os.WriteFile(tempHostsFile, []byte("127.0.0.1 localhost\n"), 0644); err != nil {
		t.Fatalf("Failed to create temp hosts file: %v", err)
	}

	m := &Manager{
		hostsPath: tempHostsFile,
		entries:   make(map[string]string),
	}

	done := make(chan bool)

	// Concurrent writers
	for i := 0; i < 5; i++ {
		go func(index int) {
			domain := "app" + string(rune('0'+index)) + ".local"
			m.Add(domain, "127.0.0.1")
			done <- true
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		go func() {
			m.List()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	entries := m.List()
	if len(entries) != 5 {
		t.Errorf("Expected 5 entries after concurrent operations, got %d", len(entries))
	}
}
