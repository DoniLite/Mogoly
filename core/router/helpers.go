package router

import (
	"github.com/DoniLite/Mogoly/core/config"
	"gopkg.in/yaml.v3"
)

func MarshalConfig(config *Config) ([]byte, error) {
	return yaml.Marshal(config)
}

func UnmarshalConfig(data []byte) (*Config, error) {
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func LoadConfig() (*Config, error) {
	data, err := config.LoadConfigFile(ROUTER_CONFIG_FILE)
	if err != nil {
		return nil, err
	}
	return UnmarshalConfig(data)
}