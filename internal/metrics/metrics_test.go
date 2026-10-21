package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestDeleteScanner(t *testing.T) {
	for _, status := range []string{"Compliant", "NonCompliant", "Skipped", "Error"} {
		CheckResults.WithLabelValues("a", "ns", status).Set(1)
		CheckResults.WithLabelValues("b", "ns", status).Set(1)
	}
	ScanTotal.WithLabelValues("a", "ns").Inc()
	ScanTotal.WithLabelValues("b", "ns").Inc()
	ScanDuration.WithLabelValues("a", "ns").Observe(1)
	ScanDuration.WithLabelValues("b", "ns").Observe(1)

	checkBefore := testutil.CollectAndCount(CheckResults, "bps_check_results")
	totalBefore := testutil.CollectAndCount(ScanTotal, "bps_scans_total")
	durationBefore := testutil.CollectAndCount(ScanDuration, "bps_scan_duration_seconds")

	DeleteScanner("a", "ns")

	if got := testutil.CollectAndCount(CheckResults, "bps_check_results"); got != checkBefore-4 {
		t.Errorf("CheckResults series = %d, want %d", got, checkBefore-4)
	}
	if got := testutil.CollectAndCount(ScanTotal, "bps_scans_total"); got != totalBefore-1 {
		t.Errorf("ScanTotal series = %d, want %d", got, totalBefore-1)
	}
	if got := testutil.CollectAndCount(ScanDuration, "bps_scan_duration_seconds"); got != durationBefore-1 {
		t.Errorf("ScanDuration series = %d, want %d", got, durationBefore-1)
	}

	// Other scanners are untouched and deleting an unknown scanner is a no-op.
	DeleteScanner("never-scanned", "ns")
	if got := testutil.ToFloat64(ScanTotal.WithLabelValues("b", "ns")); got != 1 {
		t.Errorf("scanner b ScanTotal = %v, want 1", got)
	}
}
