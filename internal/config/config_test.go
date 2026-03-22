package config

import (
	"os"
	"testing"
)

// setEnv sets an env var for the duration of the test and restores it afterwards.
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	prev, existed := os.LookupEnv(key)
	os.Setenv(key, value) //nolint:errcheck
	t.Cleanup(func() {
		if existed {
			os.Setenv(key, prev) //nolint:errcheck
		} else {
			os.Unsetenv(key) //nolint:errcheck
		}
	})
}

// unsetEnv clears an env var for the duration of the test and restores it afterwards.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	prev, existed := os.LookupEnv(key)
	os.Unsetenv(key) //nolint:errcheck
	t.Cleanup(func() {
		if existed {
			os.Setenv(key, prev) //nolint:errcheck
		}
	})
}

// TestGetEnv_Present verifies that getEnv returns the env var value when set.
func TestGetEnv_Present(t *testing.T) {
	setEnv(t, "TEST_GETENV_KEY", "hello")
	got := getEnv("TEST_GETENV_KEY", "default")
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

// TestGetEnv_Absent verifies that getEnv returns the default when the env var is not set.
func TestGetEnv_Absent(t *testing.T) {
	unsetEnv(t, "TEST_GETENV_MISSING")
	got := getEnv("TEST_GETENV_MISSING", "fallback")
	if got != "fallback" {
		t.Errorf("got %q, want %q", got, "fallback")
	}
}

// TestLoadFrom_Defaults verifies that loadFrom returns sensible defaults when no
// env vars or flags are set.
func TestLoadFrom_Defaults(t *testing.T) {
	keys := []string{"GLUETUN_URL", "INTERFACE_NAME", "LISTEN_ADDRESS", "METRICS_PATH", "LOG_LEVEL"}
	for _, k := range keys {
		unsetEnv(t, k)
	}

	cfg, err := loadFrom([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GluetunURL != "http://localhost:8000" {
		t.Errorf("GluetunURL: got %q, want %q", cfg.GluetunURL, "http://localhost:8000")
	}
	if cfg.InterfaceName != "tun0" {
		t.Errorf("InterfaceName: got %q, want %q", cfg.InterfaceName, "tun0")
	}
	if cfg.ListenAddress != ":9586" {
		t.Errorf("ListenAddress: got %q, want %q", cfg.ListenAddress, ":9586")
	}
	if cfg.MetricsPath != "/metrics" {
		t.Errorf("MetricsPath: got %q, want %q", cfg.MetricsPath, "/metrics")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel: got %q, want %q", cfg.LogLevel, "info")
	}
	if cfg.ShowVersion {
		t.Error("ShowVersion: got true, want false")
	}
}

// TestLoadFrom_EnvOverride verifies that env vars override the defaults.
func TestLoadFrom_EnvOverride(t *testing.T) {
	keys := []string{"GLUETUN_URL", "INTERFACE_NAME", "LISTEN_ADDRESS", "METRICS_PATH", "LOG_LEVEL"}
	for _, k := range keys {
		unsetEnv(t, k)
	}

	setEnv(t, "GLUETUN_URL", "http://gluetun:8000")
	setEnv(t, "INTERFACE_NAME", "wg0")
	setEnv(t, "LISTEN_ADDRESS", ":9999")
	setEnv(t, "METRICS_PATH", "/custom-metrics")
	setEnv(t, "LOG_LEVEL", "debug")

	cfg, err := loadFrom([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GluetunURL != "http://gluetun:8000" {
		t.Errorf("GluetunURL: got %q, want %q", cfg.GluetunURL, "http://gluetun:8000")
	}
	if cfg.InterfaceName != "wg0" {
		t.Errorf("InterfaceName: got %q, want %q", cfg.InterfaceName, "wg0")
	}
	if cfg.ListenAddress != ":9999" {
		t.Errorf("ListenAddress: got %q, want %q", cfg.ListenAddress, ":9999")
	}
	if cfg.MetricsPath != "/custom-metrics" {
		t.Errorf("MetricsPath: got %q, want %q", cfg.MetricsPath, "/custom-metrics")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel: got %q, want %q", cfg.LogLevel, "debug")
	}
}

// TestLoadFrom_FlagOverride verifies that CLI flags override env var defaults.
func TestLoadFrom_FlagOverride(t *testing.T) {
	keys := []string{"GLUETUN_URL", "INTERFACE_NAME", "LISTEN_ADDRESS", "METRICS_PATH", "LOG_LEVEL"}
	for _, k := range keys {
		unsetEnv(t, k)
	}

	cfg, err := loadFrom([]string{
		"--gluetun-url=http://flag-host:8000",
		"--interface-name=eth0",
		"--log-level=warn",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GluetunURL != "http://flag-host:8000" {
		t.Errorf("GluetunURL: got %q, want %q", cfg.GluetunURL, "http://flag-host:8000")
	}
	if cfg.InterfaceName != "eth0" {
		t.Errorf("InterfaceName: got %q, want %q", cfg.InterfaceName, "eth0")
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("LogLevel: got %q, want %q", cfg.LogLevel, "warn")
	}
}

// TestLoadFrom_EmptyURL verifies that an empty GLUETUN_URL returns an error.
func TestLoadFrom_EmptyURL(t *testing.T) {
	keys := []string{"GLUETUN_URL", "INTERFACE_NAME", "LISTEN_ADDRESS", "METRICS_PATH", "LOG_LEVEL"}
	for _, k := range keys {
		unsetEnv(t, k)
	}
	setEnv(t, "GLUETUN_URL", "")

	_, err := loadFrom([]string{"--gluetun-url="})
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}
