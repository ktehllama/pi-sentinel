package monitor

import (
	"fmt"
	"sort"
	"time"

	"github.com/ktehllama/pi-sentinel/config"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/process"
)

func CheckCPU(cfg config.Config) []Result {
	percents, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil || len(percents) == 0 {
		return nil
	}
	usage := percents[0]

	if usage < cfg.Thresholds.CPUPercent {
		return []Result{{
			Metric:  "CPU",
			Above:   false,
			Message: fmt.Sprintf("Back to normal (%.1f%%)", usage),
		}}
	}

	topName, topPct := topCPUProcess()
	return []Result{{
		Metric:  "CPU",
		Above:   true,
		Message: fmt.Sprintf("Usage at %.1f%% (threshold: %.1f%%) — top process: %s (%.1f%%)", usage, cfg.Thresholds.CPUPercent, topName, topPct),
	}}
}

func topCPUProcess() (string, float64) {
	procs, err := process.Processes()
	if err != nil {
		return "unknown", 0
	}
	type entry struct {
		name string
		pct  float64
	}
	var ps []entry
	for _, p := range procs {
		pct, err := p.CPUPercent()
		if err != nil || pct == 0 {
			continue
		}
		name, _ := p.Name()
		ps = append(ps, entry{name, pct})
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].pct > ps[j].pct })
	if len(ps) == 0 {
		return "unknown", 0
	}
	return ps[0].name, ps[0].pct
}
