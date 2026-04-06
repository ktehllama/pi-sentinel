package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Poll       PollConfig      `toml:"poll"`
	Thresholds ThresholdConfig `toml:"thresholds"`
	Cooldown   CooldownConfig  `toml:"cooldown"`
	Services   ServicesConfig  `toml:"services"`
	Telegram   TelegramConfig  `toml:"telegram"`
}

type PollConfig struct {
	IntervalSeconds int `toml:"interval_seconds"`
}

type ThresholdConfig struct {
	CPUPercent     float64 `toml:"cpu_percent"`
	RAMPercent     float64 `toml:"ram_percent"`
	DiskPercent    float64 `toml:"disk_percent"`
	TempCelsius    float64 `toml:"temp_celsius"`
	NetBandwidthMB float64 `toml:"net_bandwidth_mb"`
}

type CooldownConfig struct {
	AlertMinutes int `toml:"alert_minutes"`
}

type ServicesConfig struct {
	Watchlist []string `toml:"watchlist"`
}

type TelegramConfig struct {
	BotToken string `toml:"bot_token"`
	ChatID   string `toml:"chat_id"`
}

func DefaultConfig() Config {
	return Config{
		Poll: PollConfig{IntervalSeconds: 30},
		Thresholds: ThresholdConfig{
			CPUPercent:     85.0,
			RAMPercent:     80.0,
			DiskPercent:    90.0,
			TempCelsius:    75.0,
			NetBandwidthMB: 50.0,
		},
		Cooldown: CooldownConfig{AlertMinutes: 5},
		Services: ServicesConfig{Watchlist: []string{}},
		Telegram: TelegramConfig{BotToken: "", ChatID: "REDACTED"},
	}
}

func ConfigPath() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "pi-sentinel", "config.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pi-sentinel", "config.toml")
}

func Load() (Config, error) {
	cfg := DefaultConfig()
	path := ConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

func Save(cfg Config) error {
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

func Set(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	switch key {
	case "poll.interval_seconds":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid int: %s", value)
		}
		cfg.Poll.IntervalSeconds = v
	case "thresholds.cpu_percent":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float: %s", value)
		}
		cfg.Thresholds.CPUPercent = v
	case "thresholds.ram_percent":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float: %s", value)
		}
		cfg.Thresholds.RAMPercent = v
	case "thresholds.disk_percent":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float: %s", value)
		}
		cfg.Thresholds.DiskPercent = v
	case "thresholds.temp_celsius":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float: %s", value)
		}
		cfg.Thresholds.TempCelsius = v
	case "thresholds.net_bandwidth_mb":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float: %s", value)
		}
		cfg.Thresholds.NetBandwidthMB = v
	case "cooldown.alert_minutes":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid int: %s", value)
		}
		cfg.Cooldown.AlertMinutes = v
	case "services.watchlist":
		parts := strings.Split(value, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		cfg.Services.Watchlist = parts
	case "telegram.bot_token":
		cfg.Telegram.BotToken = value
	case "telegram.chat_id":
		cfg.Telegram.ChatID = value
	default:
		return fmt.Errorf("unknown config key: %s (valid keys: poll.interval_seconds, thresholds.cpu_percent, thresholds.ram_percent, thresholds.disk_percent, thresholds.temp_celsius, thresholds.net_bandwidth_mb, cooldown.alert_minutes, services.watchlist, telegram.bot_token, telegram.chat_id)", key)
	}
	return Save(cfg)
}
