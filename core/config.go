package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

func LoadConfigFIle(configPath string) ([]byte, error) {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/"
	}
	defPath := path.Join(cwd, configPath)

	content, err := os.ReadFile(defPath)

	if err != nil {
		return nil, fmt.Errorf("error during the config file reading at: %s", defPath)
	}

	return content, nil
}

func DiscoverConfigFormat(configPath string) (string, error) {
	ext := path.Ext(configPath)
	var format string

	if ext == "" {
		return "", fmt.Errorf("invalid path provided")
	}

	if strings.Contains(ext, "json") {
		format = "json"
	} else if strings.Contains(ext, "yml") {
		format = "yaml"
	}

	return format, nil
}

func ParseConfig(content []byte, format string) (*Config, error) {
	var config Config
	var err error

	if format == "json" {
		err = json.Unmarshal(content, &config)
	} else {
		err = yaml.Unmarshal(content, &content)
	}

	if err != nil {
		return nil, err
	}

	return &config, nil
}
