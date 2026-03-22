// Package config handles configuration loading from environment variables and CLI flags.
package config

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

// Config holds the application configuration.
type Config struct {
	GluetunURL    string
	InterfaceName string
	ListenAddress string
	MetricsPath   string
	LogLevel      string
	ShowVersion   bool
}

// Load parses configuration from environment variables and command-line flags.
func Load() (*Config, error) {
	return loadFrom(os.Args[1:])
}

// loadFrom parses configuration from the given argument slice.
// It is separated from Load so tests can supply their own args without
// touching os.Args or the global pflag CommandLine.
func loadFrom(args []string) (*Config, error) {
	cfg := &Config{}

	fs := pflag.NewFlagSet("gluetun-exporter", pflag.ContinueOnError)

	fs.StringVar(&cfg.GluetunURL, "gluetun-url", getEnv("GLUETUN_URL", "http://localhost:8000"),
		"Gluetun control server URL")
	fs.StringVar(&cfg.InterfaceName, "interface-name", getEnv("INTERFACE_NAME", "tun0"),
		"WireGuard interface name to read traffic stats from")
	fs.StringVar(&cfg.ListenAddress, "listen-address", getEnv("LISTEN_ADDRESS", ":9586"),
		"Address to listen on for HTTP requests")
	fs.StringVar(&cfg.MetricsPath, "metrics-path", getEnv("METRICS_PATH", "/metrics"),
		"Path under which to expose metrics")
	fs.StringVar(&cfg.LogLevel, "log-level", getEnv("LOG_LEVEL", "info"),
		"Log level (debug, info, warn, error)")
	fs.BoolVar(&cfg.ShowVersion, "version", false,
		"Print version information and exit")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parse flags: %w", err)
	}

	if cfg.GluetunURL == "" {
		return nil, fmt.Errorf("gluetun URL is required (set GLUETUN_URL or --gluetun-url)")
	}

	return cfg, nil
}

// getEnv returns the value of the environment variable key, or defaultValue if
// the variable is not set.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
