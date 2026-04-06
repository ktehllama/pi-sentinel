package monitor

// Result is returned by every monitor's Check function.
// Above is true when the metric exceeds its threshold.
// Message always describes the current state (used for both alerts and recovery).
type Result struct {
	Metric  string // e.g. "CPU", "RAM", "DISK:/", "SERVICE:nginx", "NET:eth0", "TEMP"
	Above   bool
	Message string
}
