package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhisme/tinylist/internal/config"
)

// minimalValidConfig is the smallest config.yaml that passes Validate().
const minimalValidConfig = `
server:
  public_url: "http://localhost:8080"
auth:
  username: admin
  password: secret
`

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeTempConfig: %v", err)
	}
	return path
}

// ----------------------------------------------------------------------------
// Defaults
// ----------------------------------------------------------------------------

func TestDefaultConfig_RequireVerification_DefaultsToTrue(t *testing.T) {
	cfg, err := config.LoadFromFile(writeTempConfig(t, minimalValidConfig))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Server.RequireVerification {
		t.Error("RequireVerification should default to true when not specified in config")
	}
}

func TestDefaultConfig_ExtraOrigins_DefaultsToEmpty(t *testing.T) {
	cfg, err := config.LoadFromFile(writeTempConfig(t, minimalValidConfig))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Server.ExtraOrigins) != 0 {
		t.Errorf("ExtraOrigins should default to empty, got %v", cfg.Server.ExtraOrigins)
	}
}

// ----------------------------------------------------------------------------
// require_verification flag
// ----------------------------------------------------------------------------

func TestLoadFromFile_RequireVerification_ExplicitFalse(t *testing.T) {
	path := writeTempConfig(t, `
server:
  public_url: "http://localhost:8080"
  require_verification: false
auth:
  username: admin
  password: secret
`)
	cfg, err := config.LoadFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.RequireVerification {
		t.Error("require_verification: false in config should be respected")
	}
}

func TestLoadFromFile_RequireVerification_ExplicitTrue(t *testing.T) {
	path := writeTempConfig(t, `
server:
  public_url: "http://localhost:8080"
  require_verification: true
auth:
  username: admin
  password: secret
`)
	cfg, err := config.LoadFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Server.RequireVerification {
		t.Error("require_verification: true in config should be respected")
	}
}

// ----------------------------------------------------------------------------
// extra_origins
// ----------------------------------------------------------------------------

func TestLoadFromFile_ExtraOrigins_ParsedFromConfig(t *testing.T) {
	path := writeTempConfig(t, `
server:
  public_url: "https://prod.example.com"
  extra_origins:
    - "http://localhost:5173"
    - "http://localhost:3000"
auth:
  username: admin
  password: secret
`)
	cfg, err := config.LoadFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"http://localhost:5173", "http://localhost:3000"}
	if len(cfg.Server.ExtraOrigins) != len(want) {
		t.Fatalf("ExtraOrigins: want %d entries, got %d: %v", len(want), len(cfg.Server.ExtraOrigins), cfg.Server.ExtraOrigins)
	}
	for i, w := range want {
		if cfg.Server.ExtraOrigins[i] != w {
			t.Errorf("ExtraOrigins[%d]: want %q, got %q", i, w, cfg.Server.ExtraOrigins[i])
		}
	}
}

func TestLoadFromFile_ExtraOrigins_DoesNotContainPublicURL(t *testing.T) {
	path := writeTempConfig(t, `
server:
  public_url: "https://prod.example.com"
  extra_origins:
    - "http://localhost:5173"
auth:
  username: admin
  password: secret
`)
	cfg, err := config.LoadFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, o := range cfg.Server.ExtraOrigins {
		if o == cfg.Server.PublicURL {
			t.Error("ExtraOrigins should not duplicate PublicURL")
		}
	}
}

// ----------------------------------------------------------------------------
// Optional config file (LoadFrom / Load behaviour)
// ----------------------------------------------------------------------------

func TestLoadFrom_MissingFile_SucceedsWithEnvVars(t *testing.T) {
	// Point at a path that definitely does not exist.
	missing := filepath.Join(t.TempDir(), "does-not-exist.yaml")

	t.Setenv("TINYLIST_AUTH_PASSWORD", "envpass")

	cfg, err := config.LoadFrom(missing)
	if err != nil {
		t.Fatalf("LoadFrom should succeed when file is absent and env vars satisfy requirements, got: %v", err)
	}
	if cfg.Auth.Password != "envpass" {
		t.Errorf("Auth.Password: want %q, got %q", "envpass", cfg.Auth.Password)
	}
}

func TestLoadFrom_MissingFile_NoPassword_ReturnsValidationError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.yaml")

	// Explicitly clear so neither file nor env provides a password.
	t.Setenv("TINYLIST_AUTH_PASSWORD", "")

	_, err := config.LoadFrom(missing)
	if err == nil {
		t.Error("LoadFrom should fail when password is absent from both file and env var")
	}
}

func TestLoadFrom_MissingFile_UsesDefaults(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.yaml")

	t.Setenv("TINYLIST_AUTH_PASSWORD", "envpass")

	cfg, err := config.LoadFrom(missing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Spot-check a few defaults.
	if cfg.Server.Port != 8080 {
		t.Errorf("default port: want 8080, got %d", cfg.Server.Port)
	}
	if cfg.Sending.RateLimit != 10 {
		t.Errorf("default rate_limit: want 10, got %d", cfg.Sending.RateLimit)
	}
	if !cfg.Server.RequireVerification {
		t.Error("default require_verification should be true")
	}
}

func TestLoadFrom_FileValuesTakePrecedenceOverDefaults(t *testing.T) {
	path := writeTempConfig(t, `
server:
  public_url: "https://file.example.com"
  port: 9090
auth:
  username: admin
  password: secret
`)
	cfg, err := config.LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("port from file: want 9090, got %d", cfg.Server.Port)
	}
	if cfg.Server.PublicURL != "https://file.example.com" {
		t.Errorf("public_url from file: want %q, got %q", "https://file.example.com", cfg.Server.PublicURL)
	}
}

// ----------------------------------------------------------------------------
// Env var overrides — all fields (via LoadWithPath so the file is explicit)
// ----------------------------------------------------------------------------

func TestLoadWithPath_EnvVars_OverrideFileCredentials(t *testing.T) {
	path := writeTempConfig(t, `
server:
  public_url: "http://localhost:8080"
auth:
  username: fileuser
  password: filepass
`)
	t.Setenv("TINYLIST_AUTH_USERNAME", "envuser")
	t.Setenv("TINYLIST_AUTH_PASSWORD", "envpass")

	cfg, err := config.LoadWithPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Auth.Username != "envuser" {
		t.Errorf("Auth.Username: want %q, got %q", "envuser", cfg.Auth.Username)
	}
	if cfg.Auth.Password != "envpass" {
		t.Errorf("Auth.Password: want %q, got %q", "envpass", cfg.Auth.Password)
	}
}

func TestLoadWithPath_OnlyPasswordEnvVar_OverridesPassword(t *testing.T) {
	path := writeTempConfig(t, `
server:
  public_url: "http://localhost:8080"
auth:
  username: admin
  password: filepass
`)
	t.Setenv("TINYLIST_AUTH_PASSWORD", "newpass")
	t.Setenv("TINYLIST_AUTH_USERNAME", "")

	cfg, err := config.LoadWithPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Auth.Username != "admin" {
		t.Errorf("Auth.Username: should keep file value %q, got %q", "admin", cfg.Auth.Username)
	}
	if cfg.Auth.Password != "newpass" {
		t.Errorf("Auth.Password: want %q, got %q", "newpass", cfg.Auth.Password)
	}
}

func TestLoadWithPath_EnvVarPassword_AllowsEmptyFilePassword(t *testing.T) {
	// Simulates the Kubernetes deployment pattern: auth credentials are absent
	// from the config file / ConfigMap and injected via a Secret env var instead.
	path := writeTempConfig(t, `
server:
  public_url: "http://localhost:8080"
auth:
  username: admin
  password: ""
`)
	t.Setenv("TINYLIST_AUTH_PASSWORD", "injected-secret")
	t.Setenv("TINYLIST_AUTH_USERNAME", "")

	_, err := config.LoadWithPath(path)
	if err != nil {
		t.Errorf("LoadWithPath should succeed when TINYLIST_AUTH_PASSWORD is provided via env var, got: %v", err)
	}
}

func TestLoadWithPath_NoEnvVar_EmptyFilePassword_ReturnsValidationError(t *testing.T) {
	path := writeTempConfig(t, `
server:
  public_url: "http://localhost:8080"
auth:
  username: admin
  password: ""
`)
	t.Setenv("TINYLIST_AUTH_PASSWORD", "")
	t.Setenv("TINYLIST_AUTH_USERNAME", "")

	_, err := config.LoadWithPath(path)
	if err == nil {
		t.Error("LoadWithPath should return a validation error when password is absent from both file and env var")
	}
	if err != nil && !strings.Contains(err.Error(), "password") {
		t.Errorf("validation error should mention 'password', got: %v", err)
	}
}

func TestEnvOverride_ServerHost(t *testing.T) {
	t.Setenv("TINYLIST_SERVER_HOST", "127.0.0.1")
	t.Setenv("TINYLIST_AUTH_PASSWORD", "secret")
	cfg, err := config.LoadFrom(filepath.Join(t.TempDir(), "no-file.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host: want %q, got %q", "127.0.0.1", cfg.Server.Host)
	}
}

func TestEnvOverride_ServerPort(t *testing.T) {
	t.Setenv("TINYLIST_SERVER_PORT", "9000")
	t.Setenv("TINYLIST_AUTH_PASSWORD", "secret")
	cfg, err := config.LoadFrom(filepath.Join(t.TempDir(), "no-file.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("Server.Port: want 9000, got %d", cfg.Server.Port)
	}
}

func TestEnvOverride_ServerPublicURL(t *testing.T) {
	t.Setenv("TINYLIST_SERVER_PUBLIC_URL", "https://env.example.com")
	t.Setenv("TINYLIST_AUTH_PASSWORD", "secret")
	cfg, err := config.LoadFrom(filepath.Join(t.TempDir(), "no-file.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.PublicURL != "https://env.example.com" {
		t.Errorf("Server.PublicURL: want %q, got %q", "https://env.example.com", cfg.Server.PublicURL)
	}
}

func TestEnvOverride_ServerAPIBasePath(t *testing.T) {
	t.Setenv("TINYLIST_SERVER_API_BASE_PATH", "/myapp")
	t.Setenv("TINYLIST_AUTH_PASSWORD", "secret")
	cfg, err := config.LoadFrom(filepath.Join(t.TempDir(), "no-file.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.APIBasePath != "/myapp" {
		t.Errorf("Server.APIBasePath: want %q, got %q", "/myapp", cfg.Server.APIBasePath)
	}
}

func TestEnvOverride_ServerExtraOrigins_CommaSeparated(t *testing.T) {
	t.Setenv("TINYLIST_SERVER_EXTRA_ORIGINS", "http://localhost:5173, http://localhost:3000 , http://dev.local")
	t.Setenv("TINYLIST_AUTH_PASSWORD", "secret")
	cfg, err := config.LoadFrom(filepath.Join(t.TempDir(), "no-file.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"http://localhost:5173", "http://localhost:3000", "http://dev.local"}
	if len(cfg.Server.ExtraOrigins) != len(want) {
		t.Fatalf("ExtraOrigins: want %v, got %v", want, cfg.Server.ExtraOrigins)
	}
	for i, w := range want {
		if cfg.Server.ExtraOrigins[i] != w {
			t.Errorf("ExtraOrigins[%d]: want %q, got %q", i, w, cfg.Server.ExtraOrigins[i])
		}
	}
}

func TestEnvOverride_ServerRequireVerification_False(t *testing.T) {
	t.Setenv("TINYLIST_SERVER_REQUIRE_VERIFICATION", "false")
	t.Setenv("TINYLIST_AUTH_PASSWORD", "secret")
	cfg, err := config.LoadFrom(filepath.Join(t.TempDir(), "no-file.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.RequireVerification {
		t.Error("TINYLIST_SERVER_REQUIRE_VERIFICATION=false should disable verification")
	}
}

func TestEnvOverride_ServerRequireVerification_True(t *testing.T) {
	// Default is already true; set it explicitly via env var to confirm it parses.
	t.Setenv("TINYLIST_SERVER_REQUIRE_VERIFICATION", "true")
	t.Setenv("TINYLIST_AUTH_PASSWORD", "secret")
	cfg, err := config.LoadFrom(filepath.Join(t.TempDir(), "no-file.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Server.RequireVerification {
		t.Error("TINYLIST_SERVER_REQUIRE_VERIFICATION=true should keep verification enabled")
	}
}

func TestEnvOverride_DatabasePath(t *testing.T) {
	t.Setenv("TINYLIST_DATABASE_PATH", "/data/custom.db")
	t.Setenv("TINYLIST_AUTH_PASSWORD", "secret")
	cfg, err := config.LoadFrom(filepath.Join(t.TempDir(), "no-file.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Database.Path != "/data/custom.db" {
		t.Errorf("Database.Path: want %q, got %q", "/data/custom.db", cfg.Database.Path)
	}
}

func TestEnvOverride_SendingRateLimit(t *testing.T) {
	t.Setenv("TINYLIST_SENDING_RATE_LIMIT", "50")
	t.Setenv("TINYLIST_AUTH_PASSWORD", "secret")
	cfg, err := config.LoadFrom(filepath.Join(t.TempDir(), "no-file.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Sending.RateLimit != 50 {
		t.Errorf("Sending.RateLimit: want 50, got %d", cfg.Sending.RateLimit)
	}
}

func TestEnvOverride_SendingMaxRetries(t *testing.T) {
	t.Setenv("TINYLIST_SENDING_MAX_RETRIES", "5")
	t.Setenv("TINYLIST_AUTH_PASSWORD", "secret")
	cfg, err := config.LoadFrom(filepath.Join(t.TempDir(), "no-file.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Sending.MaxRetries != 5 {
		t.Errorf("Sending.MaxRetries: want 5, got %d", cfg.Sending.MaxRetries)
	}
}

func TestEnvOverride_SendingBatchSize(t *testing.T) {
	t.Setenv("TINYLIST_SENDING_BATCH_SIZE", "250")
	t.Setenv("TINYLIST_AUTH_PASSWORD", "secret")
	cfg, err := config.LoadFrom(filepath.Join(t.TempDir(), "no-file.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Sending.BatchSize != 250 {
		t.Errorf("Sending.BatchSize: want 250, got %d", cfg.Sending.BatchSize)
	}
}

func TestEnvOverride_TakesPrecedenceOverFile(t *testing.T) {
	// File says port 9090; env var says 7777 — env var should win.
	path := writeTempConfig(t, `
server:
  port: 9090
  public_url: "http://localhost:9090"
auth:
  username: admin
  password: filepass
`)
	t.Setenv("TINYLIST_SERVER_PORT", "7777")

	cfg, err := config.LoadWithPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 7777 {
		t.Errorf("env var should override file: want 7777, got %d", cfg.Server.Port)
	}
}

// ----------------------------------------------------------------------------
// Validation
// ----------------------------------------------------------------------------

func TestLoadFromFile_MissingPassword_ReturnsError(t *testing.T) {
	path := writeTempConfig(t, `
server:
  public_url: "http://localhost:8080"
auth:
  username: admin
  password: ""
`)
	_, err := config.LoadFromFile(path)
	if err == nil {
		t.Error("LoadFromFile should return an error when auth.password is empty")
	}
}

func TestLoadFromFile_MissingUsername_ReturnsError(t *testing.T) {
	path := writeTempConfig(t, `
server:
  public_url: "http://localhost:8080"
auth:
  username: ""
  password: secret
`)
	_, err := config.LoadFromFile(path)
	if err == nil {
		t.Error("LoadFromFile should return an error when auth.username is empty")
	}
}

func TestLoadFromFile_NonExistentFile_ReturnsError(t *testing.T) {
	_, err := config.LoadFromFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("LoadFromFile should return an error for a non-existent file")
	}
}
