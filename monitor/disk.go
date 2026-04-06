package monitor

import (
	"fmt"

	"github.com/gueva/pi-sentinel/config"
	"github.com/shirou/gopsutil/v3/disk"
)

func CheckDisk(cfg config.Config) []Result {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil
	}
	var results []Result
	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		metric := "DISK:" + p.Mountpoint
		usedPct := usage.UsedPercent
		if usedPct >= cfg.Thresholds.DiskPercent {
			results = append(results, Result{
				Metric:  metric,
				Above:   true,
				Message: fmt.Sprintf("%s at %.1f%% used (threshold: %.1f%%)", p.Mountpoint, usedPct, cfg.Thresholds.DiskPercent),
			})
		} else {
			results = append(results, Result{
				Metric:  metric,
				Above:   false,
				Message: fmt.Sprintf("%s back to normal (%.1f%% used)", p.Mountpoint, usedPct),
			})
		}
	}
	return results
}
