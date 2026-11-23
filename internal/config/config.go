package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the application.
// Note: SMTP settings are configured via the admin UI and stored in the database.
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
	Host                string   `yaml:"host"`
	Port                int      `yaml:"port"`
	PublicURL           string   `yaml:"public_url"`
	APIBasePath         string   `yaml:"api_base_path"`        // Base path for API routes (e.g., "/tinylist" for /tinylist/api/*)
	ExtraOrigins        []string `yaml:"extra_origins"`        // Additional CORS origins (e.g. dev servers); never needed in production
	RequireVerification bool     `yaml:"require_verification"` // If false, subscribers are verified immediately on signup
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type SendingConfig struct {
	RateLimit  int           `yaml:"rate_limit"`  // Emails per second
	MaxRetries int           `yaml:"max_retries"` // Max retry attempts for failed sends
	RetryDelay time.Duration `yaml:"-"`           // Delay between retries (not configurable yet)
	BatchSize  int           `yaml:"batch_size"`  // Number of subscribers to process at once
}

// ---------------------------------------------------------------------------
// Public loaders
// ---------------------------------------------------------------------------

// Load is the primary entry point used by the server.
//
// Priority (highest to lowest):
//  1. Environment variables  (always applied)
//  2. config.yaml            (optional — if absent, defaults + env vars are used)
//  3. Built-in defaults
//
// This means a Docker deployment can omit the config file entirely and pass
// everything via -e / --env-file.
func Load() (*Config, error) {
	return LoadFrom("config.yaml")
}

// LoadFrom is like Load but reads from path instead of "config.yaml".
// The file is optional — if it does not exist the loader continues with
// defaults and env vars. This is useful for tests that need to control the
// file path without changing the process working directory.
func LoadFrom(path string) (*Config, error) {
	cfg := defaultConfig()
	if err := mergeFileIfExists(cfg, path); err != nil {
		return nil, err
	}
	applyEnvOverrides(cfg)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadFromFile loads and validates configuration from a specific YAML file.
// The file must exist. Env var overrides are NOT applied, making the result
// fully determined by the file — useful for tests that need deterministic config.
func LoadFromFile(path string) (*Config, error) {
	cfg := defaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadWithPath is like LoadFrom but requires the file to exist.
// Env var overrides are applied. Useful for testing env override logic against
// a known file without accepting a missing-file as valid.
func LoadWithPath(path string) (*Config, error) {
	cfg := defaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	applyEnvOverrides(cfg)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

// Validate checks that required configuration is present.
func (c *Config) Validate() error {
	if c.Auth.Password == "" {
		return fmt.Errorf("auth.password is required — set it in config.yaml or via TINYLIST_AUTH_PASSWORD")
	}
	if c.Auth.Username == "" {
		return fmt.Errorf("auth.username is required — set it in config.yaml or via TINYLIST_AUTH_USERNAME")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Internals
// ---------------------------------------------------------------------------

// mergeFileIfExists parses path into cfg. A missing file is silently ignored;
// any other read/parse error is returned.
func mergeFileIfExists(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // config file is optional when using env vars
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	return nil
}

// applyEnvOverrides overlays any TINYLIST_* environment variables onto cfg.
// Every field in config.example.yaml has a corresponding env var.
//
// Env var reference:
//
//	TINYLIST_SERVER_HOST
//	TINYLIST_SERVER_PORT
//	TINYLIST_SERVER_PUBLIC_URL
//	TINYLIST_SERVER_API_BASE_PATH
//	TINYLIST_SERVER_EXTRA_ORIGINS        comma-separated list of origins
//	TINYLIST_SERVER_REQUIRE_VERIFICATION true|false
//	TINYLIST_DATABASE_PATH
//	TINYLIST_SENDING_RATE_LIMIT
//	TINYLIST_SENDING_MAX_RETRIES
//	TINYLIST_SENDING_BATCH_SIZE
//	TINYLIST_AUTH_USERNAME
//	TINYLIST_AUTH_PASSWORD
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("TINYLIST_SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("TINYLIST_SERVER_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = n
		}
	}
	if v := os.Getenv("TINYLIST_SERVER_PUBLIC_URL"); v != "" {
		cfg.Server.PublicURL = v
	}
	if v := os.Getenv("TINYLIST_SERVER_API_BASE_PATH"); v != "" {
		cfg.Server.APIBasePath = v
	}
	if v := os.Getenv("TINYLIST_SERVER_EXTRA_ORIGINS"); v != "" {
		parts := strings.Split(v, ",")
		origins := make([]string, 0, len(parts))
		for _, p := range parts {
			if o := strings.TrimSpace(p); o != "" {
				origins = append(origins, o)
			}
		}
		cfg.Server.ExtraOrigins = origins
	}
	if v := os.Getenv("TINYLIST_SERVER_REQUIRE_VERIFICATION"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Server.RequireVerification = b
		}
	}
	if v := os.Getenv("TINYLIST_DATABASE_PATH"); v != "" {
		cfg.Database.Path = v
	}
	if v := os.Getenv("TINYLIST_SENDING_RATE_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Sending.RateLimit = n
		}
	}
	if v := os.Getenv("TINYLIST_SENDING_MAX_RETRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Sending.MaxRetries = n
		}
	}
	if v := os.Getenv("TINYLIST_SENDING_BATCH_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Sending.BatchSize = n
		}
	}
	if v := os.Getenv("TINYLIST_AUTH_USERNAME"); v != "" {
		cfg.Auth.Username = v
	}
	if v := os.Getenv("TINYLIST_AUTH_PASSWORD"); v != "" {
		cfg.Auth.Password = v
	}
}

// defaultConfig returns configuration with sensible defaults.
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:                "0.0.0.0",
			Port:                8080,
			PublicURL:           "http://localhost:8080",
			APIBasePath:         "",
			RequireVerification: true,
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
