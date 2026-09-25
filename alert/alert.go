package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ktehllama/pi-sentinel/config"
	"github.com/ktehllama/pi-sentinel/monitor"
)

var logger = log.New(os.Stdout, "", 0)

var telegramClient = &http.Client{Timeout: 10 * time.Second}

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

// shouldRecover atomically checks and clears alerting state.
// Returns true only on the alerting→normal transition.
func (d *Dispatcher) shouldRecover(r monitor.Result) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.alerting[r.Metric] {
		return false
	}
	delete(d.lastFire, r.Metric)
	d.alerting[r.Metric] = false
	return true
}

func (d *Dispatcher) printLine(level, metric, message string) {
	logger.Printf("[%s] [%s] [%s] %s",
		time.Now().Format("2006-01-02 15:04:05"), level, metric, message)
}

func sendTelegram(recovering bool, metric, message string, cfg config.Config) {
	var text string
	if recovering {
		text = fmt.Sprintf("✅ pi-sentinel [%s]\n%s", metric, message)
	} else {
		text = fmt.Sprintf("🔴 pi-sentinel [%s]\n%s", metric, message)
	}
	payload, err := json.Marshal(map[string]string{
		"chat_id": cfg.Telegram.ChatID,
		"text":    text,
	})
	if err != nil {
		logger.Printf("[%s] [ERROR] [TELEGRAM] marshal: %v",
			time.Now().Format("2006-01-02 15:04:05"), err)
		return
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.Telegram.BotToken)
	resp, err := telegramClient.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		logger.Printf("[%s] [ERROR] [TELEGRAM] %v",
			time.Now().Format("2006-01-02 15:04:05"), err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Printf("[%s] [ERROR] [TELEGRAM] status %d",
			time.Now().Format("2006-01-02 15:04:05"), resp.StatusCode)
	}
}
