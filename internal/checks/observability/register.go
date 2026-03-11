package observability

import "github.com/sebrandon1/bps-operator/internal/checks"

func init() {
	checks.Register(checks.CheckInfo{
		Name: "observability-crd-status", Category: "observability",
		Description: "Verifies CRDs define a .status subresource",
		Remediation: "Add status subresource to the CRD spec versions",
		Fn:          CheckCRDStatus,
	})
	checks.Register(checks.CheckInfo{
		Name: "observability-termination-policy", Category: "observability",
		Description: "Verifies containers set terminationMessagePolicy to FallbackToLogsOnError",
		Remediation: "Set terminationMessagePolicy to FallbackToLogsOnError",
		Fn:          CheckTerminationPolicy,
	})
}
