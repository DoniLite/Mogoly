package config

import (
	"fmt"
	"os"
)

func GetEnv(key string, defaultValue string) string {
	env := os.Getenv(fmt.Sprintf("%s%s", ENV_PREFIX, key))
	if env == "" {
		return defaultValue
	}
	return env
}

func SetEnv(key, value string) {
	os.Setenv(fmt.Sprintf("%s%s", ENV_PREFIX, key), value)
}
