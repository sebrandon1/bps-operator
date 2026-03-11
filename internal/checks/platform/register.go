package platform

import "github.com/sebrandon1/bps-operator/internal/checks"

func init() {
	checks.Register(checks.CheckInfo{
		Name: "platform-boot-params", Category: "platform",
		Description: "Verifies no non-standard kernel boot parameters are set",
		Remediation: "Use MachineConfig to manage kernel boot parameters",
		Fn:          CheckBootParams,
	})
	checks.Register(checks.CheckInfo{
		Name: "platform-hugepages", Category: "platform",
		Description: "Verifies hugepage configuration is consistent",
		Remediation: "Configure hugepages via MachineConfig or performance profile",
		Fn:          CheckHugepages,
	})
	checks.Register(checks.CheckInfo{
		Name: "platform-sysctl", Category: "platform",
		Description: "Verifies sysctl settings are not modified outside MachineConfig",
		Remediation: "Use MachineConfig to manage sysctl settings",
		Fn:          CheckSysctl,
	})
	checks.Register(checks.CheckInfo{
		Name: "platform-tainted", Category: "platform",
		Description: "Verifies the kernel is not tainted",
		Remediation: "Investigate and resolve kernel taint causes",
		Fn:          CheckTainted,
	})
}
