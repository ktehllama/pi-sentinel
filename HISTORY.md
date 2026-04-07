# Pi Sentinel — Project History

## Overview

Pi Sentinel is a lightweight Go daemon that monitors a Raspberry Pi's system health and sends alerts via stdout and Telegram DM when anomalies are detected. It is designed to be stealthy — minimal RAM and CPU footprint — while covering all major resource types.

**Binary:** single static Go binary, cross-compiled for Linux ARM64  
**Deploy:** runs as a `systemctl` service on the Pi (`gueva@192.168.1.45`)  
**Config:** `~/.config/pi-sentinel/config.toml`, editable live via `pi-sentinel set <key> <value>`

---

## What It Monitors

| Metric | Source | What It Detects |
|---|---|---|
| CPU | gopsutil/cpu + process | Overall % + top offending process |
| Memory | gopsutil/mem + process | RAM % used + top process by RSS |
| Disk | gopsutil/disk | Per-partition % used |
| Services | exec systemctl | Watchlisted service not active/running |
| Network | gopsutil/net (delta) | Per-interface MB/s over threshold |
| Temperature | /sys/class/thermal/thermal_zone0/temp | CPU temp °C |

---

## Default Thresholds

| Metric | Default |
|---|---|
| CPU | 85% |
| RAM | 80% |
| Disk | 90% |
| Temperature | 75°C |
| Network | 50 MB/s per interface |
| Cooldown (alert repeat) | 5 minutes |

---

## Architecture

```
pi-sentinel/
├── main.go              # entry point — daemon + set subcommands
├── config/config.go     # load/save/set config.toml
├── monitor/
│   ├── types.go         # shared Result struct
│   ├── cpu.go
│   ├── memory.go
│   ├── disk.go
│   ├── services.go
│   ├── network.go
│   └── temperature.go
├── alert/alert.go       # dispatcher: cooldown, recovery, stdout + Telegram
└── build.sh             # cross-compile + SCP to Pi
```

**Alert lifecycle:** each monitor returns `[]Result{Metric, Above bool, Message}`. The `alert.Dispatcher` tracks per-metric alerting state — fires a `WARN` when a metric crosses its threshold (respecting cooldown), and fires an `INFO` recovery when it drops back below.

---

## CLI Usage

```bash
# Start monitoring loop (run by systemctl)
pi-sentinel daemon

# Live-edit a threshold (no restart needed)
pi-sentinel set thresholds.cpu_percent 90
pi-sentinel set services.watchlist nginx,myapp,redis
pi-sentinel set cooldown.alert_minutes 10
```

---

## Deployment

```bash
# From Windows dev machine:
bash build.sh   # cross-compiles ARM64 binary + SCPs to Pi

# On the Pi:
sudo systemctl enable pi-sentinel
sudo systemctl start pi-sentinel
sudo journalctl -u pi-sentinel -f   # live logs
```

---

## Changelog

### 2026-04-06 — v1.0.0 (initial release)

- Project scaffolded: Go module `github.com/gueva/pi-sentinel`, gopsutil v3, BurntSushi/toml
- Config package: TOML-based config with defaults, XDG-aware path, live `set` command
- Alert dispatcher: stateful cooldown + recovery detection, stdout + Telegram DM
- 6 monitors: CPU, RAM, disk, services watchlist, network bandwidth (stateful delta), temperature
- Main daemon loop: re-reads config each tick, graceful SIGTERM/SIGINT shutdown
- Build script: ARM64 cross-compile with `-ldflags="-s -w"`, SCP to Pi
- Pushed to GitHub: `github.com/ktehllama/pi-sentinel`
