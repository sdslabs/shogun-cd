package config

import (
	"fmt"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"go.yaml.in/yaml/v3"
)

func setConfigPath() string {
	if _, err := os.Stat("/etc/shogun/config.yaml"); err == nil {
		return "/etc/shogun/config.yaml"
	}
	return "config.yaml"
}

func LoadConfigs(l utils.Logger) (*Config, error) {
	configFilePath := setConfigPath()
	l.Log("Loading Configs from %v...", configFilePath)
	var cfg Config

	f, err := os.Open(configFilePath)
	if err != nil {
		l.LogNewError("Failed to open config file:", err)
		return nil, err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		l.LogNewError("Failed to decode config file:", err)
		return nil, err
	}

	if cfg.Timezone == "" {
		cfg.Timezone = "UTC"
	}
	if _, err := time.LoadLocation(cfg.Timezone); err != nil {
		return nil, fmt.Errorf("invalid timezone: %w", err)
	}
	l.Log("Configs loaded successfully")

	// [TODO] Validate config here if needed

	return &cfg, nil
}
