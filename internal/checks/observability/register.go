package observability

import "github.com/sebrandon1/bps-operator/internal/checks"

func init() {
	checks.Register(checks.CheckInfo{
		Name: "observability-crd-status", Category: "observability",
		Description: "Verifies CRDs define a .status subresource",
		Remediation: "Add status subresource to the CRD spec versions",
		CatalogID:   "observability-crd-status",
		Fn:          CheckCRDStatus,
	})
	checks.Register(checks.CheckInfo{
		Name: "observability-termination-policy", Category: "observability",
		Description: "Verifies containers set terminationMessagePolicy to FallbackToLogsOnError",
		Remediation: "Set terminationMessagePolicy to FallbackToLogsOnError",
		CatalogID:   "observability-termination-policy",
		Fn:          CheckTerminationPolicy,
	})
	checks.Register(checks.CheckInfo{
		Name: "observability-pod-disruption-budget", Category: "observability",
		Description: "Verifies PodDisruptionBudgets exist for HA workloads",
		Remediation: "Create a PodDisruptionBudget for deployments with replicas > 1",
		CatalogID:   "observability-pod-disruption-budget",
		Fn:          CheckPodDisruptionBudget,
	})
}
