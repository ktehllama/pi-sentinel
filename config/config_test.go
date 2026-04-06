package config

import (
	"testing"
)

func setTempHome(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Poll.IntervalSeconds != 30 {
		t.Errorf("expected interval 30, got %d", cfg.Poll.IntervalSeconds)
	}
	if cfg.Thresholds.CPUPercent != 85.0 {
		t.Errorf("expected cpu threshold 85.0, got %f", cfg.Thresholds.CPUPercent)
	}
	if cfg.Thresholds.RAMPercent != 80.0 {
		t.Errorf("expected ram threshold 80.0, got %f", cfg.Thresholds.RAMPercent)
	}
	if cfg.Cooldown.AlertMinutes != 5 {
		t.Errorf("expected cooldown 5, got %d", cfg.Cooldown.AlertMinutes)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	setTempHome(t)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Poll.IntervalSeconds != 30 {
		t.Errorf("expected default interval 30, got %d", cfg.Poll.IntervalSeconds)
	}
}

func TestSaveAndLoad(t *testing.T) {
	setTempHome(t)
	cfg := DefaultConfig()
	cfg.Thresholds.CPUPercent = 70.0
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Thresholds.CPUPercent != 70.0 {
		t.Errorf("expected 70.0, got %f", loaded.Thresholds.CPUPercent)
	}
}

func TestSetFloat(t *testing.T) {
	setTempHome(t)
	if err := Set("thresholds.cpu_percent", "92.5"); err != nil {
		t.Fatal(err)
	}
	cfg, _ := Load()
	if cfg.Thresholds.CPUPercent != 92.5 {
		t.Errorf("expected 92.5, got %f", cfg.Thresholds.CPUPercent)
	}
}

func TestSetWatchlist(t *testing.T) {
	setTempHome(t)
	if err := Set("services.watchlist", "nginx,redis,postgresql"); err != nil {
		t.Fatal(err)
	}
	cfg, _ := Load()
	if len(cfg.Services.Watchlist) != 3 {
		t.Errorf("expected 3 services, got %d", len(cfg.Services.Watchlist))
	}
	if cfg.Services.Watchlist[0] != "nginx" {
		t.Errorf("expected nginx, got %s", cfg.Services.Watchlist[0])
	}
}

func TestSetInterval(t *testing.T) {
	setTempHome(t)
	if err := Set("poll.interval_seconds", "60"); err != nil {
		t.Fatal(err)
	}
	cfg, _ := Load()
	if cfg.Poll.IntervalSeconds != 60 {
		t.Errorf("expected 60, got %d", cfg.Poll.IntervalSeconds)
	}
}

func TestSetUnknownKeyReturnsError(t *testing.T) {
	setTempHome(t)
	if err := Set("invalid.key", "value"); err == nil {
		t.Error("expected error for unknown key")
	}
}
