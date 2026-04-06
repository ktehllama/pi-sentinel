package monitor

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/gueva/pi-sentinel/config"
)

func CheckServices(cfg config.Config) []Result {
	var results []Result
	for _, svc := range cfg.Services.Watchlist {
		metric := "SERVICE:" + svc
		out, err := exec.Command("systemctl", "is-active", svc).Output()
		status := strings.TrimSpace(string(out))
		if err != nil || status != "active" {
			results = append(results, Result{
				Metric:  metric,
				Above:   true,
				Message: fmt.Sprintf("%s is %s (expected: active)", svc, status),
			})
		} else {
			results = append(results, Result{
				Metric:  metric,
				Above:   false,
				Message: fmt.Sprintf("%s is active", svc),
			})
		}
	}
	return results
}
