package config

import (
	"fmt"
	"os"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	APIKey    string
	Tenant    string
	Namespace string
}

// Load reads configuration from environment variables.
// Returns an error if F5XC_API_KEY is not set (required for CLI and export modes).
func Load() (*Config, error) {
	cfg := Defaults()
	cfg.APIKey = os.Getenv("F5XC_API_KEY")
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("F5XC_API_KEY environment variable is required")
	}
	return cfg, nil
}

// Defaults returns a Config populated from env vars with fallback values.
// APIKey may be empty — intended for web server mode where the key is supplied via the UI.
func Defaults() *Config {
	cfg := &Config{
		Tenant:    os.Getenv("F5XC_TENANT"),
		Namespace: os.Getenv("F5XC_NAMESPACE"),
	}
	if cfg.Tenant == "" {
		cfg.Tenant = "f5-sa"
	}
	if cfg.Namespace == "" {
		cfg.Namespace = "s-iannetta"
	}
	return cfg
}
