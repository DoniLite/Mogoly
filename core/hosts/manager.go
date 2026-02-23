// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package hosts provides cross-platform hosts file management.
// It allows adding, removing, and listing domain-to-IP mappings
// in the system hosts file (/etc/hosts on Unix, Windows equivalent).
package hosts

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/DoniLite/Mogoly/core/config"
)

const MogolyMarker string = "# Managed by Mogoly"

// Manager handles hosts file operations
type Manager struct {
	hostsPath string
	entries   map[string]string // domain -> IP
	mu        sync.RWMutex
}

// NewManager creates a new hosts file manager
func NewManager() (*Manager, error) {
	path := getHostsFilePath()
	if envPath := config.GetEnv(config.MOGOLY_HOSTS_PATH, ""); envPath != "" {
		path = envPath
	}

	hm := &Manager{
		hostsPath: path,
		entries:   make(map[string]string),
	}

	// Load existing Mogoly-managed entries
	if err := hm.loadEntries(); err != nil {
		return nil, err
	}

	return hm, nil
}

// Add adds a domain -> IP mapping to the hosts file
func (m *Manager) Add(domain string, ip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate inputs
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	if ip == "" {
		return fmt.Errorf("IP address cannot be empty")
	}

	// Check if entry already exists with same IP
	if existingIP, exists := m.entries[domain]; exists && existingIP == ip {
		return nil // Already correct
	}

	// Read current hosts file
	content, err := os.ReadFile(m.hostsPath)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied: hosts file modification requires administrator privileges\n\nPlease run with sudo:\n  sudo mogoly domain add %s --local", domain)
		}
		return fmt.Errorf("failed to read hosts file: %v", err)
	}

	// Parse lines and check for existing domain
	lines := strings.Split(string(content), getLineEnding())
	domainExists := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		parts := strings.Fields(trimmed)
		if len(parts) >= 2 {
			if slices.Contains(parts[1:], domain) {
				// Update existing entry
				lines[i] = fmt.Sprintf("%s %s %s", ip, domain, MogolyMarker)
				domainExists = true
			}
		}
	}

	// Add new entry if doesn't exist
	if !domainExists {
		newEntry := fmt.Sprintf("%s %s %s", ip, domain, MogolyMarker)
		lines = append(lines, newEntry)
	}

	// Write back to hosts file
	newContent := strings.Join(lines, getLineEnding())
	if !strings.HasSuffix(newContent, getLineEnding()) {
		newContent += getLineEnding()
	}

	if err := os.WriteFile(m.hostsPath, []byte(newContent), 0644); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied: hosts file modification requires administrator privileges\n\nPlease run with sudo:\n  sudo mogoly domain add %s --local", domain)
		}
		return fmt.Errorf("failed to write hosts file: %v", err)
	}

	m.entries[domain] = ip
	return nil
}

// Remove removes a domain from the hosts file
func (m *Manager) Remove(domain string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.entries[domain]; !exists {
		return nil // Already removed
	}

	content, err := os.ReadFile(m.hostsPath)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied: hosts file modification requires administrator privileges")
		}
		return fmt.Errorf("failed to read hosts file: %v", err)
	}

	lines := strings.Split(string(content), getLineEnding())
	var newLines []string

	for _, line := range lines {
		// Keep lines that are not Mogoly-managed entries for this domain
		if strings.Contains(line, MogolyMarker) && strings.Contains(line, domain) {
			continue // Skip this line
		}
		newLines = append(newLines, line)
	}

	newContent := strings.Join(newLines, getLineEnding())
	if err := os.WriteFile(m.hostsPath, []byte(newContent), 0644); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied: hosts file modification requires administrator privileges")
		}
		return fmt.Errorf("failed to write hosts file: %v", err)
	}

	delete(m.entries, domain)
	return nil
}

// Has checks if a domain exists in the hosts file
func (m *Manager) Has(domain string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.entries[domain]
	return exists
}

// Get returns the IP for a domain
func (m *Manager) Get(domain string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ip, exists := m.entries[domain]
	return ip, exists
}

// List returns all Mogoly-managed entries
func (m *Manager) List() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := make(map[string]string)
	maps.Copy(entries, m.entries)
	return entries
}

// loadEntries loads existing Mogoly-managed entries from hosts file
func (m *Manager) loadEntries() error {
	file, err := os.Open(m.hostsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Hosts file doesn't exist, start fresh
		}
		return fmt.Errorf("failed to open hosts file: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Only load Mogoly-managed entries
		if !strings.Contains(line, MogolyMarker) {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			ip := parts[0]
			domain := parts[1]
			m.entries[domain] = ip
		}
	}

	return scanner.Err()
}

// GetHostsPath returns the current hosts file path
func (m *Manager) GetHostsPath() string {
	return m.hostsPath
}
