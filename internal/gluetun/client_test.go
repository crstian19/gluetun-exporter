package gluetun_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crstian19/gluetun-exporter/internal/gluetun"
)

func newTestClient(t *testing.T, handler http.Handler) (*gluetun.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return gluetun.NewClient(srv.URL), srv
}

// TestGetVPNStatus_Running verifies that a 200 response with status "running" is parsed correctly.
func TestGetVPNStatus_Running(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/vpn/status" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"running"}`))
	}))

	status, err := client.GetVPNStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Status != "running" {
		t.Errorf("Status: got %q, want %q", status.Status, "running")
	}
}

// TestGetVPNStatus_NonOK verifies that a non-200 response returns an error.
func TestGetVPNStatus_NonOK(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))

	_, err := client.GetVPNStatus(context.Background())
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
}

// TestGetPublicIP_Success verifies that the full public IP JSON is parsed correctly.
func TestGetPublicIP_Success(t *testing.T) {
	body := `{
		"public_ip":    "1.2.3.4",
		"country":      "Netherlands",
		"region":       "Noord-Holland",
		"city":         "Amsterdam",
		"organization": "AS1234 Example BV",
		"hostname":     "vpn.example.com",
		"timezone":     "Europe/Amsterdam"
	}`
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))

	ip, err := client.GetPublicIP(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip.IP != "1.2.3.4" {
		t.Errorf("IP: got %q, want %q", ip.IP, "1.2.3.4")
	}
	if ip.Country != "Netherlands" {
		t.Errorf("Country: got %q, want %q", ip.Country, "Netherlands")
	}
	if ip.Region != "Noord-Holland" {
		t.Errorf("Region: got %q, want %q", ip.Region, "Noord-Holland")
	}
	if ip.City != "Amsterdam" {
		t.Errorf("City: got %q, want %q", ip.City, "Amsterdam")
	}
	if ip.Organization != "AS1234 Example BV" {
		t.Errorf("Organization: got %q, want %q", ip.Organization, "AS1234 Example BV")
	}
	if ip.Hostname != "vpn.example.com" {
		t.Errorf("Hostname: got %q, want %q", ip.Hostname, "vpn.example.com")
	}
	if ip.Timezone != "Europe/Amsterdam" {
		t.Errorf("Timezone: got %q, want %q", ip.Timezone, "Europe/Amsterdam")
	}
}

// TestGetPortForwarding_Success verifies that a forwarded port is parsed correctly.
func TestGetPortForwarding_Success(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"port":12345}`))
	}))

	pf, err := client.GetPortForwarding(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pf.Port != 12345 {
		t.Errorf("Port: got %d, want 12345", pf.Port)
	}
}

// TestGetPortForwarding_NonOK verifies that a non-200 response returns an error.
func TestGetPortForwarding_NonOK(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))

	_, err := client.GetPortForwarding(context.Background())
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
}
