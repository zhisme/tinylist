package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	SMTP     SMTPConfig     `yaml:"smtp"`
	Sending  SendingConfig  `yaml:"sending"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	PublicURL string `yaml:"public_url"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type SMTPConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	FromEmail string `yaml:"from_email"`
	FromName  string `yaml:"from_name"`
	TLS       bool   `yaml:"tls"`
}

type SendingConfig struct {
	RateLimit  int           `yaml:"rate_limit"`   // Emails per second
	MaxRetries int           `yaml:"max_retries"`  // Max retry attempts for failed sends
	RetryDelay time.Duration `yaml:"-"`            // Delay between retries (parsed from seconds)
	BatchSize  int           `yaml:"batch_size"`   // Number of subscribers to process at once
}

type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json or text
}

// Load loads configuration from YAML file (if exists) and environment variables
// Environment variables override YAML values
func Load() (*Config, error) {
	return LoadFromFile("config.yaml")
}

// LoadFromFile loads configuration from a specific YAML file
func LoadFromFile(path string) (*Config, error) {
	// Start with defaults
	cfg := defaultConfig()

	// Try to load from YAML file
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables
	applyEnvOverrides(cfg)

	return cfg, nil
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
		SMTP: SMTPConfig{
			Host:      "",
			Port:      587,
			Username:  "",
			Password:  "",
			FromEmail: "",
			FromName:  "Newsletter",
			TLS:       true,
		},
		Sending: SendingConfig{
			RateLimit:  10,
			MaxRetries: 3,
			RetryDelay: 5 * time.Second,
			BatchSize:  100,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

// applyEnvOverrides applies environment variable overrides to config
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("TINYLIST_SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("TINYLIST_SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("TINYLIST_PUBLIC_URL"); v != "" {
		cfg.Server.PublicURL = v
	}

	if v := os.Getenv("TINYLIST_DATABASE_PATH"); v != "" {
		cfg.Database.Path = v
	}

	if v := os.Getenv("TINYLIST_SMTP_HOST"); v != "" {
		cfg.SMTP.Host = v
	}
	if v := os.Getenv("TINYLIST_SMTP_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.SMTP.Port = port
		}
	}
	if v := os.Getenv("TINYLIST_SMTP_USERNAME"); v != "" {
		cfg.SMTP.Username = v
	}
	if v := os.Getenv("TINYLIST_SMTP_PASSWORD"); v != "" {
		cfg.SMTP.Password = v
	}
	if v := os.Getenv("TINYLIST_SMTP_FROM_EMAIL"); v != "" {
		cfg.SMTP.FromEmail = v
	}
	if v := os.Getenv("TINYLIST_SMTP_FROM_NAME"); v != "" {
		cfg.SMTP.FromName = v
	}
	if v := os.Getenv("TINYLIST_SMTP_TLS"); v != "" {
		if tls, err := strconv.ParseBool(v); err == nil {
			cfg.SMTP.TLS = tls
		}
	}

	if v := os.Getenv("TINYLIST_SENDING_RATE_LIMIT"); v != "" {
		if rate, err := strconv.Atoi(v); err == nil {
			cfg.Sending.RateLimit = rate
		}
	}
	if v := os.Getenv("TINYLIST_SENDING_MAX_RETRIES"); v != "" {
		if retries, err := strconv.Atoi(v); err == nil {
			cfg.Sending.MaxRetries = retries
		}
	}
	if v := os.Getenv("TINYLIST_SENDING_RETRY_DELAY_SECONDS"); v != "" {
		if seconds, err := strconv.Atoi(v); err == nil {
			cfg.Sending.RetryDelay = time.Duration(seconds) * time.Second
		}
	}
	if v := os.Getenv("TINYLIST_SENDING_BATCH_SIZE"); v != "" {
		if batch, err := strconv.Atoi(v); err == nil {
			cfg.Sending.BatchSize = batch
		}
	}

	if v := os.Getenv("TINYLIST_LOGGING_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("TINYLIST_LOGGING_FORMAT"); v != "" {
		cfg.Logging.Format = v
	}
}

