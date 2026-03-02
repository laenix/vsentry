package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the collector configuration
type Config struct {
	Name     string         `yaml:"name"`
	Type     string         `yaml:"type"`
	Interval int            `yaml:"interval"`
	Channels []string       `yaml:"channels"`        // For Windows (comma-separated)
	Sources  []SourceConfig `yaml:"sources"`        // For Linux
	Ingest   IngestConfig   `yaml:"ingest"`
	Hostname string         `yaml:"hostname,omitempty"`
}

// SourceConfig defines a single log source
type SourceConfig struct {
	Type    string `yaml:"type"`     // syslog, nginx_access, ssh, etc
	Path    string `yaml:"path"`     // file path
	Format  string `yaml:"format"`   // log format
	Enabled bool   `yaml:"enabled"`  // whether to collect
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