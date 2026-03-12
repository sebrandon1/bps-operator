package platform

import "github.com/sebrandon1/bps-operator/internal/checks"

func init() {
	checks.Register(checks.CheckInfo{
		Name: "platform-boot-params", Category: "platform",
		Description: "Verifies no non-standard kernel boot parameters are set",
		Remediation: "Use MachineConfig to manage kernel boot parameters",
		CatalogID:   "platform-alteration-boot-params",
		Fn:          CheckBootParams,
	})
	checks.Register(checks.CheckInfo{
		Name: "platform-hugepages", Category: "platform",
		Description: "Verifies hugepage configuration is consistent",
		Remediation: "Configure hugepages via MachineConfig or performance profile",
		CatalogID:   "platform-alteration-hugepages-config",
		Fn:          CheckHugepages,
	})
	checks.Register(checks.CheckInfo{
		Name: "platform-sysctl", Category: "platform",
		Description: "Verifies sysctl settings are not modified outside MachineConfig",
		Remediation: "Use MachineConfig to manage sysctl settings",
		CatalogID:   "platform-alteration-sysctl-config",
		Fn:          CheckSysctl,
	})
	checks.Register(checks.CheckInfo{
		Name: "platform-tainted", Category: "platform",
		Description: "Verifies the kernel is not tainted",
		Remediation: "Investigate and resolve kernel taint causes",
		CatalogID:   "platform-alteration-tainted-node-kernel",
		Fn:          CheckTainted,
	})

	// Spec-based platform checks
	checks.Register(checks.CheckInfo{
		Name: "platform-service-mesh-usage", Category: "platform",
		Description: "Verifies pods do not use service mesh (Istio) sidecars",
		Remediation: "Remove Istio sidecar injection if not required",
		CatalogID:   "platform-alteration-service-mesh-usage",
		Fn:          CheckServiceMeshUsage,
	})
	checks.Register(checks.CheckInfo{
		Name: "platform-hugepages-2mi-only", Category: "platform",
		Description: "Verifies only 2Mi hugepages are used (not 1Gi)",
		Remediation: "Use 2Mi hugepages instead of 1Gi",
		CatalogID:   "platform-alteration-hugepages-2mi-only",
		Fn:          CheckHugepages2MiOnly,
	})
	checks.Register(checks.CheckInfo{
		Name: "platform-ocp-node-count", Category: "platform",
		Description: "Verifies cluster has minimum recommended number of worker nodes",
		Remediation: "Ensure cluster has at least 3 worker nodes",
		CatalogID:   "platform-alteration-ocp-node-count",
		Fn:          CheckNodeCount,
	})
}
