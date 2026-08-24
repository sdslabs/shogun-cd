package config

import (
	"os"

	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"go.yaml.in/yaml/v3"
)

var configFilePath = "config.yaml"

func LoadConfigs(l utils.Logger) (*Config, error) {
	l.Log("Loading Configs from %v...", configFilePath)
	var cfg Config

	raw, err := os.ReadFile(configFilePath)
	if err != nil {
		l.LogNewError("Failed to open config file:", err)
		return nil, err
	}

	// Expand ${VAR}/$VAR references so container envs (e.g. from docker-compose) can fill in the config.
	expanded := os.ExpandEnv(string(raw))

	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		l.LogNewError("Failed to decode config file:", err)
		return nil, err
	}

	l.Log("Configs loaded successfully")

	// [TODO] Validate config here if needed

	return &cfg, nil
}
