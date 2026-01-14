package daemon

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	// Unix socket paths
	UnixSocketPath = "/var/run/mogoly.sock"
	UserSocketPath = ".mogoly/mogoly.sock"

	// Windows named pipe
	WindowsPipePath = `\\.\pipe\mogoly`

	// State and config directories
	ConfigDir = ".mogoly"
	StateFile = "daemon.json"
	PIDFile   = "mogoly.pid"
	LogFile   = "mogoly.log"
)

// GetSocketPath returns the appropriate socket path for the current platform
func GetSocketPath() string {
	if runtime.GOOS == "windows" {
		return WindowsPipePath
	}

	// Check if running as root
	if os.Getuid() == 0 {
		return UnixSocketPath
	}

	// Use user home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/tmp", "mogoly.sock")
	}

	return filepath.Join(home, UserSocketPath)
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ConfigDir)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

// GetStateFilePath returns the path to the daemon state file
func GetStateFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, StateFile), nil
}

// GetPIDFilePath returns the path to the PID file
func GetPIDFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, PIDFile), nil
}

// GetLogFilePath returns the path to the log file
func GetLogFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, LogFile), nil
}
