// Package collector implements the Prometheus collector for Gluetun metrics.
package collector

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/crstian19/gluetun-exporter/internal/gluetun"
)

const defaultProcNetDev = "/proc/net/dev"

// GluetunCollector implements the prometheus.Collector interface.
type GluetunCollector struct {
	client        *gluetun.Client
	interfaceName string

	// VPN metrics
	vpnConnected *prometheus.Desc
	publicIPInfo *prometheus.Desc

	// Port forwarding (OpenVPN providers only)
	forwardedPort *prometheus.Desc

	// Network traffic metrics (read from /proc/net/dev)
	networkRxBytes   *prometheus.Desc
	networkTxBytes   *prometheus.Desc
	networkRxPackets *prometheus.Desc
	networkTxPackets *prometheus.Desc
	networkRxErrors  *prometheus.Desc
	networkTxErrors  *prometheus.Desc
	networkRxDropped *prometheus.Desc
	networkTxDropped *prometheus.Desc

	// Exporter self-metrics
	scrapeDuration *prometheus.Desc
	scrapeErrors   prometheus.Counter
}

// NewGluetunCollector creates a new GluetunCollector.
func NewGluetunCollector(client *gluetun.Client, interfaceName string) *GluetunCollector {
	iface := []string{"interface"}
	return &GluetunCollector{
		client:        client,
		interfaceName: interfaceName,

		vpnConnected: prometheus.NewDesc(
			"gluetun_vpn_connected",
			"1 if the VPN tunnel is connected, 0 otherwise.",
			nil, nil,
		),
		publicIPInfo: prometheus.NewDesc(
			"gluetun_public_ip_info",
			"Public IP information (always 1, details in labels).",
			[]string{"ip", "country", "region", "city", "organization", "timezone"},
			nil,
		),
		forwardedPort: prometheus.NewDesc(
			"gluetun_forwarded_port",
			"Forwarded port number (1 if active, 0 if none; OpenVPN providers only).",
			[]string{"port"},
			nil,
		),
		networkRxBytes: prometheus.NewDesc(
			"gluetun_network_received_bytes_total",
			"Total bytes received through the VPN interface.",
			iface, nil,
		),
		networkTxBytes: prometheus.NewDesc(
			"gluetun_network_transmitted_bytes_total",
			"Total bytes transmitted through the VPN interface.",
			iface, nil,
		),
		networkRxPackets: prometheus.NewDesc(
			"gluetun_network_received_packets_total",
			"Total packets received through the VPN interface.",
			iface, nil,
		),
		networkTxPackets: prometheus.NewDesc(
			"gluetun_network_transmitted_packets_total",
			"Total packets transmitted through the VPN interface.",
			iface, nil,
		),
		networkRxErrors: prometheus.NewDesc(
			"gluetun_network_received_errors_total",
			"Total receive errors on the VPN interface.",
			iface, nil,
		),
		networkTxErrors: prometheus.NewDesc(
			"gluetun_network_transmitted_errors_total",
			"Total transmit errors on the VPN interface.",
			iface, nil,
		),
		networkRxDropped: prometheus.NewDesc(
			"gluetun_network_received_dropped_total",
			"Total received packets dropped on the VPN interface.",
			iface, nil,
		),
		networkTxDropped: prometheus.NewDesc(
			"gluetun_network_transmitted_dropped_total",
			"Total transmitted packets dropped on the VPN interface.",
			iface, nil,
		),
		scrapeDuration: prometheus.NewDesc(
			"gluetun_exporter_scrape_duration_seconds",
			"Duration of the scrape in seconds.",
			nil, nil,
		),
		scrapeErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gluetun_exporter_scrape_errors_total",
			Help: "Total number of scrape errors.",
		}),
	}
}

// Describe implements prometheus.Collector.
func (c *GluetunCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.vpnConnected
	ch <- c.publicIPInfo
	ch <- c.forwardedPort
	ch <- c.networkRxBytes
	ch <- c.networkTxBytes
	ch <- c.networkRxPackets
	ch <- c.networkTxPackets
	ch <- c.networkRxErrors
	ch <- c.networkTxErrors
	ch <- c.networkRxDropped
	ch <- c.networkTxDropped
	ch <- c.scrapeDuration
	c.scrapeErrors.Describe(ch)
}

// Collect implements prometheus.Collector.
func (c *GluetunCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// VPN status
	status, err := c.client.GetVPNStatus(ctx)
	if err != nil {
		slog.Error("Failed to get VPN status", "error", err)
		c.scrapeErrors.Inc()
		ch <- prometheus.MustNewConstMetric(c.vpnConnected, prometheus.GaugeValue, 0)
	} else {
		connected := 0.0
		if status.Status == "running" {
			connected = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.vpnConnected, prometheus.GaugeValue, connected)
	}

	// Public IP info
	publicIP, err := c.client.GetPublicIP(ctx)
	if err != nil {
		slog.Warn("Failed to get public IP info", "error", err)
	} else if publicIP.IP != "" {
		ch <- prometheus.MustNewConstMetric(
			c.publicIPInfo,
			prometheus.GaugeValue,
			1,
			publicIP.IP, publicIP.Country, publicIP.Region, publicIP.City, publicIP.Organization, publicIP.Timezone,
		)
	}

	// Port forwarding (OpenVPN providers only — silently skip for WireGuard)
	pf, err := c.client.GetPortForwarding(ctx)
	if err != nil {
		slog.Debug("Port forwarding not available", "error", err)
	} else if pf.Port > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.forwardedPort,
			prometheus.GaugeValue,
			1,
			strconv.Itoa(int(pf.Port)),
		)
	}

	// Network interface stats from /proc/net/dev
	stats, err := readInterfaceStats(c.interfaceName, defaultProcNetDev)
	if err != nil {
		slog.Warn("Failed to read interface stats", "interface", c.interfaceName, "error", err)
		c.scrapeErrors.Inc()
	} else {
		iface := c.interfaceName
		ch <- prometheus.MustNewConstMetric(c.networkRxBytes, prometheus.CounterValue, float64(stats.RxBytes), iface)
		ch <- prometheus.MustNewConstMetric(c.networkTxBytes, prometheus.CounterValue, float64(stats.TxBytes), iface)
		ch <- prometheus.MustNewConstMetric(c.networkRxPackets, prometheus.CounterValue, float64(stats.RxPackets), iface)
		ch <- prometheus.MustNewConstMetric(c.networkTxPackets, prometheus.CounterValue, float64(stats.TxPackets), iface)
		ch <- prometheus.MustNewConstMetric(c.networkRxErrors, prometheus.CounterValue, float64(stats.RxErrors), iface)
		ch <- prometheus.MustNewConstMetric(c.networkTxErrors, prometheus.CounterValue, float64(stats.TxErrors), iface)
		ch <- prometheus.MustNewConstMetric(c.networkRxDropped, prometheus.CounterValue, float64(stats.RxDropped), iface)
		ch <- prometheus.MustNewConstMetric(c.networkTxDropped, prometheus.CounterValue, float64(stats.TxDropped), iface)
	}

	ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, time.Since(start).Seconds())
	c.scrapeErrors.Collect(ch)
}

// interfaceStats holds network counters for a single interface.
type interfaceStats struct {
	RxBytes   uint64
	TxBytes   uint64
	RxPackets uint64
	TxPackets uint64
	RxErrors  uint64
	TxErrors  uint64
	RxDropped uint64
	TxDropped uint64
}

// readInterfaceStats reads counters for a given interface from the given procPath
// (normally /proc/net/dev).
//
// /proc/net/dev column layout (after "iface:"):
//
//	[0]  rx_bytes  [1]  rx_packets [2]  rx_errs [3]  rx_drop
//	[4]  rx_fifo   [5]  rx_frame   [6]  rx_comp  [7]  rx_multicast
//	[8]  tx_bytes  [9]  tx_packets [10] tx_errs  [11] tx_drop
//	[12] tx_fifo   [13] tx_colls   [14] tx_carrier [15] tx_comp
func readInterfaceStats(ifaceName, procPath string) (*interfaceStats, error) {
	f, err := os.Open(procPath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", procPath, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			slog.Warn("Failed to close /proc/net/dev", "error", cerr)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, ifaceName+":") {
			continue
		}

		fields := strings.Fields(strings.TrimPrefix(line, ifaceName+":"))
		if len(fields) < 16 {
			return nil, fmt.Errorf("unexpected format for interface %s", ifaceName)
		}

		parse := func(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) }

		var s interfaceStats
		var parseErr error
		if s.RxBytes, parseErr = parse(fields[0]); parseErr != nil {
			return nil, fmt.Errorf("parse rx_bytes: %w", parseErr)
		}
		if s.RxPackets, parseErr = parse(fields[1]); parseErr != nil {
			return nil, fmt.Errorf("parse rx_packets: %w", parseErr)
		}
		if s.RxErrors, parseErr = parse(fields[2]); parseErr != nil {
			return nil, fmt.Errorf("parse rx_errors: %w", parseErr)
		}
		if s.RxDropped, parseErr = parse(fields[3]); parseErr != nil {
			return nil, fmt.Errorf("parse rx_dropped: %w", parseErr)
		}
		if s.TxBytes, parseErr = parse(fields[8]); parseErr != nil {
			return nil, fmt.Errorf("parse tx_bytes: %w", parseErr)
		}
		if s.TxPackets, parseErr = parse(fields[9]); parseErr != nil {
			return nil, fmt.Errorf("parse tx_packets: %w", parseErr)
		}
		if s.TxErrors, parseErr = parse(fields[10]); parseErr != nil {
			return nil, fmt.Errorf("parse tx_errors: %w", parseErr)
		}
		if s.TxDropped, parseErr = parse(fields[11]); parseErr != nil {
			return nil, fmt.Errorf("parse tx_dropped: %w", parseErr)
		}

		return &s, nil
	}

	return nil, fmt.Errorf("interface %s not found in /proc/net/dev", ifaceName)
}
