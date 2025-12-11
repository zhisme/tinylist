package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the application
// Note: SMTP settings are configured via admin UI and stored in database
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Sending  SendingConfig  `yaml:"sending"`
	Auth     AuthConfig     `yaml:"auth"`
}

type AuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type ServerConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	PublicURL string `yaml:"public_url"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type SendingConfig struct {
	RateLimit  int           `yaml:"rate_limit"`  // Emails per second
	MaxRetries int           `yaml:"max_retries"` // Max retry attempts for failed sends
	RetryDelay time.Duration `yaml:"-"`           // Delay between retries (parsed from seconds)
	BatchSize  int           `yaml:"batch_size"`  // Number of subscribers to process at once
}

// Load loads configuration from YAML file
func Load() (*Config, error) {
	return LoadFromFile("config.yaml")
}

// LoadFromFile loads configuration from a specific YAML file
func LoadFromFile(path string) (*Config, error) {
	// Start with defaults
	cfg := defaultConfig()

	// Config file is required
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file (required): %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that required configuration is present
func (c *Config) Validate() error {
	if c.Auth.Password == "" {
		return fmt.Errorf("auth.password is required - admin panel cannot run without authentication")
	}
	if c.Auth.Username == "" {
		return fmt.Errorf("auth.username is required")
	}
	return nil
}

// defaultConfig returns configuration with sensible defaults
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:      "0.0.0.0",
			Port:      8080,
			PublicURL: "http://localhost:8080",
		},
		Database: DatabaseConfig{
			Path: "./data/tinylist.db",
		},
		Sending: SendingConfig{
			RateLimit:  10,
			MaxRetries: 3,
			RetryDelay: 5 * time.Second,
			BatchSize:  100,
		},
		Auth: AuthConfig{
			Username: "admin",
			Password: "",
		},
	}
}
