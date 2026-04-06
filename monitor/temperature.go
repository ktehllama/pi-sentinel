package monitor

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gueva/pi-sentinel/config"
)

const thermalPath = "/sys/class/thermal/thermal_zone0/temp"

func CheckTemperature(cfg config.Config) []Result {
	data, err := os.ReadFile(thermalPath)
	if err != nil {
		return nil
	}
	millideg, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return nil
	}
	tempC := float64(millideg) / 1000.0

	if tempC >= cfg.Thresholds.TempCelsius {
		return []Result{{
			Metric:  "TEMP",
			Above:   true,
			Message: fmt.Sprintf("CPU temp %.1f°C (threshold: %.1f°C)", tempC, cfg.Thresholds.TempCelsius),
		}}
	}
	return []Result{{
		Metric:  "TEMP",
		Above:   false,
		Message: fmt.Sprintf("CPU temp normal (%.1f°C)", tempC),
	}}
}
