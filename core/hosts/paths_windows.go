//go:build windows
// +build windows

// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package hosts

import (
	"os"
	"path/filepath"
)

// getHostsFilePath returns the Windows hosts file path
func getHostsFilePath() string {
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemRoot = "C:\\Windows"
	}
	return filepath.Join(systemRoot, "System32", "drivers", "etc", "hosts")
}

// getLineEnding returns the Windows line ending
func getLineEnding() string {
	return "\r\n"
}
