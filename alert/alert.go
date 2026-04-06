package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gueva/pi-sentinel/config"
	"github.com/gueva/pi-sentinel/monitor"
)

// Dispatcher tracks alert state and dispatches to stdout + Telegram.
type Dispatcher struct {
	mu       sync.Mutex
	alerting map[string]bool
	lastFire map[string]time.Time
}

func New() *Dispatcher {
	return &Dispatcher{
		alerting: make(map[string]bool),
		lastFire: make(map[string]time.Time),
	}
}

// Process evaluates a monitor Result and fires alerts or recovery messages as needed.
func (d *Dispatcher) Process(r monitor.Result, cfg config.Config) {
	if r.Above {
		if d.shouldFire(r, cfg) {
			d.printLine("WARN", r.Metric, r.Message)
			if cfg.Telegram.BotToken != "" {
				go sendTelegram(false, r.Metric, r.Message, cfg)
			}
		}
	} else {
		if d.shouldRecover(r) {
			d.mu.Lock()
			delete(d.lastFire, r.Metric)
			d.alerting[r.Metric] = false
			d.mu.Unlock()
			d.printLine("INFO", r.Metric, r.Message)
			if cfg.Telegram.BotToken != "" {
				go sendTelegram(true, r.Metric, r.Message, cfg)
			}
		}
	}
}

// shouldFire returns true and updates state if the alert should be dispatched.
func (d *Dispatcher) shouldFire(r monitor.Result, cfg config.Config) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	cooldown := time.Duration(cfg.Cooldown.AlertMinutes) * time.Minute

	last, fired := d.lastFire[r.Metric]
	if fired && now.Sub(last) < cooldown {
		return false
	}
	d.alerting[r.Metric] = true
	d.lastFire[r.Metric] = now
	return true
}

// shouldRecover returns true if the metric was previously alerting (transition to normal).
func (d *Dispatcher) shouldRecover(r monitor.Result) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.alerting[r.Metric]
}

func (d *Dispatcher) printLine(level, metric, message string) {
	fmt.Printf("[%s] [%s] [%s] %s\n",
		time.Now().Format("2006-01-02 15:04:05"), level, metric, message)
}

func sendTelegram(recovering bool, metric, message string, cfg config.Config) {
	var text string
	if recovering {
		text = fmt.Sprintf("✅ pi-sentinel [%s]\n%s", metric, message)
	} else {
		text = fmt.Sprintf("🔴 pi-sentinel [%s]\n%s", metric, message)
	}
	payload, _ := json.Marshal(map[string]string{
		"chat_id": cfg.Telegram.ChatID,
		"text":    text,
	})
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.Telegram.BotToken)
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		fmt.Printf("[%s] [ERROR] [TELEGRAM] %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		return
	}
	defer resp.Body.Close()
}
