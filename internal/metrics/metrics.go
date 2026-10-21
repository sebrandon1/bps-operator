package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	ScanDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bps_scan_duration_seconds",
			Help:    "Duration of best-practice scans in seconds.",
			Buckets: prometheus.ExponentialBuckets(1, 2, 10), // 1s, 2s, 4s, ..., 512s
		},
		[]string{"scanner", "namespace"},
	)

	ScanTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bps_scans_total",
			Help: "Total number of completed scans.",
		},
		[]string{"scanner", "namespace"},
	)

	CheckResults = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bps_check_results",
			Help: "Number of check results by status from the most recent scan.",
		},
		[]string{"scanner", "namespace", "status"},
	)

	CheckDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bps_check_duration_seconds",
			Help:    "Duration of individual best-practice checks in seconds.",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 12), // 10ms, 20ms, ..., ~40s
		},
		[]string{"check", "category"},
	)
)

func init() {
	metrics.Registry.MustRegister(ScanDuration, ScanTotal, CheckResults, CheckDuration)
}

// DeleteScanner removes every per-scanner series for the given scanner so a
// deleted scanner stops being reported. Safe to call for scanners that never
// produced metrics.
func DeleteScanner(name, namespace string) {
	labels := prometheus.Labels{"scanner": name, "namespace": namespace}
	ScanDuration.DeletePartialMatch(labels)
	ScanTotal.DeletePartialMatch(labels)
	CheckResults.DeletePartialMatch(labels)
}
