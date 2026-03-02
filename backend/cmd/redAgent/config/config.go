package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the collector configuration
type Config struct {
	Name     string      `yaml:"name"`
	Type     string      `yaml:"type"`
	Channels []string    `yaml:"channels"`
	Ingest   IngestConfig `yaml:"ingest"`
	Interval int         `yaml:"interval"`
	Hostname string      `yaml:"hostname,omitempty"`
}

// IngestConfig contains VSentry ingestion settings
type IngestConfig struct {
	Endpoint    string `yaml:"endpoint"`
	Token       string `yaml:"token"`
	StreamFields string `yaml:"stream_fields"`
}

// Load reads configuration from YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if cfg.Interval == 0 {
		cfg.Interval = 5
	}
	if cfg.Hostname == "" {
		cfg.Hostname, _ = os.Hostname()
	}

	return &cfg, nil
}