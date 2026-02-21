//go:build !windows
// +build !windows

package daemon

// isWindows returns false on Unix-like systems
func isWindows() bool {
	return false
}
