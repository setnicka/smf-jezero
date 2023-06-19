package config

import (
	"encoding/json"
	"os"
)

// ServerConfig holds configuration for the HTTP server
type ServerConfig struct {
	OrgLogin    string `json:"org_login"`
	OrgPassword string `json:"org_password"`

	StaticDir   string `json:"static_dir"`
	TemplateDir string `json:"template_dir"`

	SessionSecret string `json:"session_secret"`
	SessionMaxAge int    `json:"session_max_age"` // seconds

	Listen string `json:"listen"`
}

// Config holds all the configuration
type Config struct {
	Server ServerConfig `json:"server"`
}

// Load configuration from given file
func Load(filename string) (*Config, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}

	// TODO: checks

	return &cfg, nil
}
