package alert

import (
	"testing"
	"time"

	"github.com/gueva/pi-sentinel/config"
	"github.com/gueva/pi-sentinel/monitor"
)

func testCfg() config.Config {
	cfg := config.DefaultConfig()
	cfg.Cooldown.AlertMinutes = 60 // long cooldown by default in tests
	cfg.Telegram.BotToken = ""     // disable Telegram in tests
	return cfg
}

func TestFirstAlertFires(t *testing.T) {
	d := New()
	cfg := testCfg()
	r := monitor.Result{Metric: "CPU", Above: true, Message: "Usage at 90%"}
	fired := d.shouldFire(r, cfg)
	if !fired {
		t.Error("expected first alert to fire")
	}
}

func TestCooldownBlocksRepeat(t *testing.T) {
	d := New()
	cfg := testCfg()
	r := monitor.Result{Metric: "CPU", Above: true, Message: "Usage at 90%"}

	d.shouldFire(r, cfg) // first — fires, sets cooldown

	fired := d.shouldFire(r, cfg) // second — should be blocked
	if fired {
		t.Error("expected cooldown to block second alert")
	}
}

func TestCooldownExpiry(t *testing.T) {
	d := New()
	cfg := testCfg()
	cfg.Cooldown.AlertMinutes = 0
	r := monitor.Result{Metric: "CPU", Above: true, Message: "Usage at 90%"}

	d.shouldFire(r, cfg)
	// Backdate last fire so cooldown is expired
	d.mu.Lock()
	d.lastFire["CPU"] = time.Now().Add(-2 * time.Minute)
	d.mu.Unlock()

	fired := d.shouldFire(r, cfg)
	if !fired {
		t.Error("expected alert to fire after cooldown expiry")
	}
}

func TestRecoveryFires(t *testing.T) {
	d := New()
	// Set alerting state
	d.mu.Lock()
	d.alerting["RAM"] = true
	d.mu.Unlock()

	r := monitor.Result{Metric: "RAM", Above: false, Message: "Back to normal (42%)"}
	recovered := d.shouldRecover(r)
	if !recovered {
		t.Error("expected recovery to fire when metric was alerting")
	}
}

func TestRecoveryDoesNotFireIfNotAlerting(t *testing.T) {
	d := New()
	r := monitor.Result{Metric: "RAM", Above: false, Message: "Back to normal (42%)"}
	recovered := d.shouldRecover(r)
	if recovered {
		t.Error("expected no recovery when metric was not alerting")
	}
}

func TestRecoveryClearsCooldown(t *testing.T) {
	d := New()
	cfg := testCfg()

	// Simulate previous alerting state
	d.mu.Lock()
	d.alerting["DISK:/"] = true
	d.lastFire["DISK:/"] = time.Now()
	d.mu.Unlock()

	r := monitor.Result{Metric: "DISK:/", Above: false, Message: "Back to normal (60%)"}
	d.Process(r, cfg)

	d.mu.Lock()
	_, hasLastFire := d.lastFire["DISK:/"]
	isAlerting := d.alerting["DISK:/"]
	d.mu.Unlock()

	if hasLastFire {
		t.Error("expected lastFire to be cleared after recovery")
	}
	if isAlerting {
		t.Error("expected alerting to be false after recovery")
	}
}
