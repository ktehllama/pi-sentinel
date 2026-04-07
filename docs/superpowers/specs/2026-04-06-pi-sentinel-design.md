# Pi Sentinel — Design Spec

## Architecture

Single Go binary: `pi-sentinel`. Two subcommands:
- `pi-sentinel daemon` — monitoring loop, run by systemctl
- `pi-sentinel set <key> <value>` — live-edit config without restarting daemon

Poll loop runs every 30s (configurable). Each tick runs 6 checks sequentially,
compares against thresholds, dispatches alerts with per-metric cooldown (default 5 min).
Recovery messages fire when a metric returns below threshold.

## Metrics

| Check | Source | Detects |
|---|---|---|
| CPU | gopsutil/cpu + process | Overall % + top offending process |
| Memory | gopsutil/mem + process | RAM % used + top process by RSS |
| Disk | gopsutil/disk | Per-partition % used |
| Services | exec systemctl | Watchlisted service not active/running |
| Network | gopsutil/net (delta) | Per-interface MB/s over threshold |
| Temperature | /sys/class/thermal/thermal_zone0/temp | CPU temp °C |

## Config (~/.config/pi-sentinel/config.toml)

Auto-created on first run. Defaults baked in, overridden by file, editable via CLI.

## Alert Format

Stdout: `[2006-01-02 15:04:05] [WARN] [CPU] Usage at 91.2% — top process: ffmpeg (78%)`
Telegram: `🔴 pi-sentinel [CPU]\nUsage at 91.2% — top: ffmpeg`
Recovery: `✅ pi-sentinel [CPU]\nBack to normal (42%)`

## Deployment

Cross-compiled on Windows (GOOS=linux GOARCH=arm64), SCP to Pi, installed as systemctl service.
