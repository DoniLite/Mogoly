//go:build !windows
// +build !windows

// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package hosts

// getHostsFilePath returns the Unix hosts file path
func getHostsFilePath() string {
	return "/etc/hosts"
}

// getLineEnding returns the Unix line ending
func getLineEnding() string {
	return "\n"
}
