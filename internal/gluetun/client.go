// Package gluetun provides an HTTP client for the Gluetun control server API.
package gluetun

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultTimeout = 10 * time.Second

// Client is a Gluetun control server HTTP client
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Gluetun client
func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    baseURL,
	}
}

// VPNStatus represents the VPN connection status
type VPNStatus struct {
	Status string `json:"status"`
}

// PublicIP represents the public IP information from Gluetun
type PublicIP struct {
	IP           string `json:"public_ip"`
	Country      string `json:"country"`
	Region       string `json:"region"`
	City         string `json:"city"`
	Organization string `json:"organization"`
	Hostname     string `json:"hostname"`
	Timezone     string `json:"timezone"`
}

// GetVPNStatus fetches the current VPN connection status
func (c *Client) GetVPNStatus(ctx context.Context) (*VPNStatus, error) {
	var status VPNStatus
	if err := c.get(ctx, "/v1/vpn/status", &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// GetPublicIP fetches the current public IP and location info
func (c *Client) GetPublicIP(ctx context.Context) (*PublicIP, error) {
	var ip PublicIP
	if err := c.get(ctx, "/v1/publicip/ip", &ip); err != nil {
		return nil, err
	}
	return &ip, nil
}

// PortForwarding represents the forwarded port from Gluetun (OpenVPN providers)
type PortForwarding struct {
	Port uint16 `json:"port"`
}

// GetPortForwarding fetches the forwarded port (only available with supported OpenVPN providers)
func (c *Client) GetPortForwarding(ctx context.Context) (*PortForwarding, error) {
	var pf PortForwarding
	if err := c.get(ctx, "/v1/openvpn/portforwarded", &pf); err != nil {
		return nil, err
	}
	return &pf, nil
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, path)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}
