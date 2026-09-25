package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ktehllama/pi-sentinel/alert"
	"github.com/ktehllama/pi-sentinel/config"
	"github.com/ktehllama/pi-sentinel/monitor"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "daemon":
		runDaemon()
	case "set":
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "usage: pi-sentinel set <key> <value>")
			os.Exit(1)
		}
		if err := config.Set(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("set %s = %s\n", os.Args[2], os.Args[3])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("usage: pi-sentinel <command>")
	fmt.Println("  daemon          start monitoring loop")
	fmt.Println("  set <key> <val> update a config value")
}

func runDaemon() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Write config file on first run so user can find and edit it
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not write config: %v\n", err)
	}

	fmt.Printf("[%s] pi-sentinel started (poll every %ds, config: %s)\n",
		time.Now().Format("2006-01-02 15:04:05"),
		cfg.Poll.IntervalSeconds,
		config.ConfigPath())

	dispatcher := alert.New()
	ticker := time.NewTicker(time.Duration(cfg.Poll.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Run first tick immediately
	tick(dispatcher)

	for {
		select {
		case <-ticker.C:
			// Re-read config each tick so live edits take effect
			cfg, _ = config.Load()
			// Reset ticker if interval changed
			ticker.Reset(time.Duration(cfg.Poll.IntervalSeconds) * time.Second)
			tick(dispatcher)
		case <-quit:
			fmt.Printf("[%s] pi-sentinel stopped\n", time.Now().Format("2006-01-02 15:04:05"))
			return
		}
	}
}

func tick(d *alert.Dispatcher) {
	cfg, err := config.Load()
	if err != nil {
		return
	}
	var results []monitor.Result
	results = append(results, monitor.CheckCPU(cfg)...)
	results = append(results, monitor.CheckMemory(cfg)...)
	results = append(results, monitor.CheckDisk(cfg)...)
	results = append(results, monitor.CheckServices(cfg)...)
	results = append(results, monitor.CheckNetwork(cfg)...)
	results = append(results, monitor.CheckTemperature(cfg)...)

	for _, r := range results {
		d.Process(r, cfg)
	}
}
