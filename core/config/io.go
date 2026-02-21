package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	BASE_CONFIG_DIR string = ".mogoly"
)

func buildPathFromHome(path ...string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}
	parts := append([]string{homeDir, BASE_CONFIG_DIR}, path...)
	return filepath.Join(parts...), nil
}

func CreateConfigFile(configPath string) (string, error) {
	path, err := buildPathFromHome(configPath)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		return "", err
	}
	return path, nil
}

func CreateConfigDir(configDir string) (string, error) {
	path, err := buildPathFromHome(configDir)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}
	return path, nil
}

func LoadConfigFile(configPath string) ([]byte, error) {
	path, err := buildPathFromHome(configPath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)	
}