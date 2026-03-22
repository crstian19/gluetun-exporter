# Gluetun Exporter

<div align="center">

<img src="gluetun-exporter-logo.png" alt="Gluetun Exporter" width="180" />

**Prometheus exporter for [Gluetun](https://github.com/qdm12/gluetun) — VPN status, public IP, and network traffic metrics**

![CI](https://img.shields.io/github/actions/workflow/status/crstian19/gluetun-exporter/ci.yml?label=CI&logo=github&logoColor=white&style=for-the-badge&branch=main)
![Go Report Card](https://goreportcard.com/badge/github.com/crstian19/gluetun-exporter?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/github/license/crstian19/gluetun-exporter?style=for-the-badge&logo=unlicense&logoColor=white)
![Docker](https://img.shields.io/badge/docker-ghcr.io-blue?style=for-the-badge&logo=docker&logoColor=white)
![GitHub Release](https://img.shields.io/github/release/crstian19/gluetun-exporter?style=for-the-badge&logo=github&logoColor=white&label=release)
![Go Version](https://img.shields.io/github/go-mod/go-version/crstian19/gluetun-exporter?style=for-the-badge&logo=go&logoColor=white&label=go)

[Quick Start](#-quick-start) • [Metrics](#-metrics) • [Configuration](#%EF%B8%8F-configuration) • [Development](#%EF%B8%8F-development)

</div>

---

## 📋 Overview

Exposes Gluetun's VPN status and network traffic via the Gluetun control server API and `/proc/net/dev`.

- 🔌 **VPN status** — connected/disconnected gauge
- 🌍 **Public IP info** — IP, country, city, org via labels
- 📊 **Network traffic** — RX/TX bytes, packets, errors, drops from the VPN interface
- 🔀 **Port forwarding** — forwarded port gauge (OpenVPN providers)
- 🐳 **Multi-arch Docker** — `linux/amd64` and `linux/arm64`

---

## 🚀 Quick Start

### Docker

```bash
docker run -d \
  --name gluetun-exporter \
  --network container:gluetun \
  -p 9586:9586 \
  ghcr.io/crstian19/gluetun-exporter:latest
```

### Prometheus scrape config

```yaml
scrape_configs:
  - job_name: gluetun
    static_configs:
      - targets: ['gluetun-exporter:9586']
```

---

## 📊 Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `gluetun_vpn_connected` | Gauge | 1 if VPN tunnel is up, 0 otherwise | — |
| `gluetun_public_ip_info` | Gauge | Always 1, IP info in labels | ip, country, region, city, organization, timezone |
| `gluetun_forwarded_port` | Gauge | 1 if port is forwarded (OpenVPN providers only) | port |
| `gluetun_network_received_bytes_total` | Counter | Bytes received through VPN interface | interface |
| `gluetun_network_transmitted_bytes_total` | Counter | Bytes transmitted through VPN interface | interface |
| `gluetun_network_received_packets_total` | Counter | Packets received through VPN interface | interface |
| `gluetun_network_transmitted_packets_total` | Counter | Packets transmitted through VPN interface | interface |
| `gluetun_network_received_errors_total` | Counter | Receive errors on VPN interface | interface |
| `gluetun_network_transmitted_errors_total` | Counter | Transmit errors on VPN interface | interface |
| `gluetun_network_received_dropped_total` | Counter | Received packets dropped on VPN interface | interface |
| `gluetun_network_transmitted_dropped_total` | Counter | Transmitted packets dropped on VPN interface | interface |
| `gluetun_exporter_scrape_duration_seconds` | Gauge | Scrape duration | — |
| `gluetun_exporter_scrape_errors_total` | Counter | Total scrape errors | — |

---

## ⚙️ Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `GLUETUN_URL` | `http://localhost:8000` | Gluetun control server URL |
| `INTERFACE_NAME` | `tun0` | VPN network interface name |
| `LISTEN_ADDRESS` | `:9586` | Exporter listen address |
| `METRICS_PATH` | `/metrics` | Metrics endpoint path |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

All options also available as CLI flags — run `gluetun-exporter --help` for details.

---

## 🏗️ Development

```bash
# Build
go build -o gluetun-exporter .

# Test
go test -v -race ./...

# Lint
golangci-lint run

# Docker
docker build -t gluetun-exporter .
```

**Project structure:**

```
.
├── main.go
├── internal/
│   ├── collector/     # Prometheus collector + /proc/net/dev parsing
│   ├── gluetun/       # Gluetun control server HTTP client
│   └── config/        # Configuration
├── Dockerfile
└── .github/workflows/ # CI/CD
```

---

## 📄 License

MIT — see [LICENSE](LICENSE)

---

<div align="center">

**⭐ If this project helped you, consider giving it a star!**

Made with ❤️ from 🇪🇸

</div>
