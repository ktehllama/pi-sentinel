package monitor

import (
	"fmt"
	"sort"

	"github.com/ktehllama/pi-sentinel/config"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

func CheckMemory(cfg config.Config) []Result {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil
	}
	usedPct := vm.UsedPercent

	if usedPct < cfg.Thresholds.RAMPercent {
		return []Result{{
			Metric:  "RAM",
			Above:   false,
			Message: fmt.Sprintf("Back to normal (%.1f%%)", usedPct),
		}}
	}

	topName, topMB := topRAMProcess()
	return []Result{{
		Metric:  "RAM",
		Above:   true,
		Message: fmt.Sprintf("Usage at %.1f%% (threshold: %.1f%%) — top process: %s (%.0f MB)", usedPct, cfg.Thresholds.RAMPercent, topName, topMB),
	}}
}

func topRAMProcess() (string, float64) {
	procs, err := process.Processes()
	if err != nil {
		return "unknown", 0
	}
	type entry struct {
		name string
		rss  uint64
	}
	var ps []entry
	for _, p := range procs {
		mi, err := p.MemoryInfo()
		if err != nil || mi == nil {
			continue
		}
		name, _ := p.Name()
		ps = append(ps, entry{name, mi.RSS})
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].rss > ps[j].rss })
	if len(ps) == 0 {
		return "unknown", 0
	}
	return ps[0].name, float64(ps[0].rss) / 1024 / 1024
}
