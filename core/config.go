// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/DoniLite/Mogoly/core/router"
	"github.com/DoniLite/Mogoly/core/server"
	"gopkg.in/yaml.v3"
)

func LoadConfigFile(configPath string) ([]byte, error) {
	content, err := os.ReadFile(configPath)

	if err != nil {
		return nil, fmt.Errorf("error during the config file reading at: %s \n %v", configPath, err)
	}

	return content, nil
}

func DiscoverConfigFormat(configPath string) (string, error) {
	ext := strings.ToLower(path.Ext(configPath))
	switch ext {
	case ".json":
		return "json", nil
	case ".yml", ".yaml":
		return "yaml", nil
	case "":
		return "", fmt.Errorf("invalid path provided (no extension)")
	default:
		return "", fmt.Errorf("unsupported config extension: %s", ext)
	}
}

func ParseConfig(content []byte, format string) (*router.Config, error) {
	var config router.Config
	switch format {
	case "json":
		if err := json.Unmarshal(content, &config); err != nil {
			return nil, err
		}
	case "yaml":
		if err := yaml.Unmarshal(content, &config); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
	return &config, nil
}

func SerializeHealthCheckStatus(status *server.HealthCheckStatus) (string, error) {
	b, err := json.Marshal(status)

	if err != nil {
		return "", err
	}

	return string(b), nil
}
