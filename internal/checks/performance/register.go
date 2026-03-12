package performance

import "github.com/sebrandon1/bps-operator/internal/checks"

func init() {
	checks.Register(checks.CheckInfo{
		Name: "performance-exclusive-cpu-pool", Category: "performance",
		Description: "Verifies containers requesting whole CPUs use exclusive CPU pool (Guaranteed QoS)",
		Remediation: "Set CPU requests equal to limits with whole-number values for exclusive CPU allocation",
		CatalogID:   "performance-exclusive-cpu-pool",
		Fn:          CheckExclusiveCPUPool,
	})
	checks.Register(checks.CheckInfo{
		Name: "performance-rt-apps-no-exec-probes", Category: "performance",
		Description: "Verifies real-time containers do not use exec probes",
		Remediation: "Use httpGet or tcpSocket probes instead of exec for RT workloads",
		CatalogID:   "performance-rt-apps-no-exec-probes",
		Fn:          CheckRTAppsNoExecProbes,
	})
	checks.Register(checks.CheckInfo{
		Name: "performance-limit-memory-allocation", Category: "performance",
		Description: "Verifies containers have memory limits set",
		Remediation: "Set resources.limits.memory on all containers",
		CatalogID:   "performance-limit-memory-allocation",
		Fn:          CheckMemoryLimit,
	})
}
