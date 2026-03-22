// Package main is the entry point for the gluetun-exporter binary.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/crstian19/gluetun-exporter/internal/collector"
	"github.com/crstian19/gluetun-exporter/internal/config"
	"github.com/crstian19/gluetun-exporter/internal/gluetun"
)

// Version, GitCommit, and BuildDate are injected at build time via -ldflags.
var (
	Version   = "dev"     //nolint:gochecknoglobals
	GitCommit = "none"    //nolint:gochecknoglobals
	BuildDate = "unknown" //nolint:gochecknoglobals
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	logLevel := parseLogLevel(cfg.LogLevel)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	if cfg.ShowVersion {
		fmt.Printf("gluetun-exporter\nVersion:    %s\nGit Commit: %s\nBuild Date: %s\n", Version, GitCommit, BuildDate)
		os.Exit(0)
	}

	gluetunClient := gluetun.NewClient(cfg.GluetunURL)
	col := collector.NewGluetunCollector(gluetunClient, cfg.InterfaceName)
	prometheus.MustRegister(col)

	mux := http.NewServeMux()
	mux.Handle(cfg.MetricsPath, promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Gluetun Exporter</title></head>
<body>
<h1>Gluetun Prometheus Exporter</h1>
<p>Version: %s | Commit: %s | Built: %s</p>
<p><a href="%s">Metrics</a> | <a href="/health">Health</a></p>
<h2>Metrics exposed</h2>
<ul>
  <li>gluetun_vpn_connected - 1 if VPN tunnel is up</li>
  <li>gluetun_public_ip_info - Public IP, country, city</li>
  <li>gluetun_network_received_bytes_total - RX bytes via VPN interface</li>
  <li>gluetun_network_transmitted_bytes_total - TX bytes via VPN interface</li>
</ul>
</body>
</html>`, Version, GitCommit, BuildDate, cfg.MetricsPath)
	})

	server := &http.Server{
		Addr:         cfg.ListenAddress,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("Starting gluetun-exporter",
			"version", Version,
			"listen_address", cfg.ListenAddress,
			"metrics_path", cfg.MetricsPath,
			"gluetun_url", cfg.GluetunURL,
			"interface", cfg.InterfaceName,
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("Shutting down gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Error during shutdown", "error", err)
	}
	slog.Info("Exporter stopped")
}

// parseLogLevel converts a level string to a slog.Level value.
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
