package monitor

import (
	"fmt"
	"sync"
	"time"

	"github.com/ktehllama/pi-sentinel/config"
	psnet "github.com/shirou/gopsutil/v3/net"
)

var (
	netMu       sync.Mutex
	prevNetIO   map[string]psnet.IOCountersStat
	prevNetTime time.Time
)

func CheckNetwork(cfg config.Config) []Result {
	counters, err := psnet.IOCounters(true)
	if err != nil {
		return nil
	}

	now := time.Now()
	netMu.Lock()
	defer netMu.Unlock()

	if prevNetIO == nil {
		prevNetIO = make(map[string]psnet.IOCountersStat)
		for _, c := range counters {
			prevNetIO[c.Name] = c
		}
		prevNetTime = now
		return nil
	}

	elapsed := now.Sub(prevNetTime).Seconds()
	if elapsed <= 0 {
		return nil
	}

	var results []Result
	for _, c := range counters {
		prev, ok := prevNetIO[c.Name]
		if !ok {
			prevNetIO[c.Name] = c
			continue
		}
		rxMB := float64(c.BytesRecv-prev.BytesRecv) / elapsed / 1024 / 1024
		txMB := float64(c.BytesSent-prev.BytesSent) / elapsed / 1024 / 1024
		totalMB := rxMB + txMB
		metric := "NET:" + c.Name

		if totalMB >= cfg.Thresholds.NetBandwidthMB {
			results = append(results, Result{
				Metric:  metric,
				Above:   true,
				Message: fmt.Sprintf("%s bandwidth %.1f MB/s (rx: %.1f, tx: %.1f, threshold: %.1f)", c.Name, totalMB, rxMB, txMB, cfg.Thresholds.NetBandwidthMB),
			})
		} else {
			results = append(results, Result{
				Metric:  metric,
				Above:   false,
				Message: fmt.Sprintf("%s bandwidth normal (%.1f MB/s)", c.Name, totalMB),
			})
		}
		prevNetIO[c.Name] = c
	}
	prevNetTime = now
	return results
}
