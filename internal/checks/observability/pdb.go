package observability

import (
	"fmt"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// CheckPodDisruptionBudget verifies PodDisruptionBudgets exist for HA workloads.
func CheckPodDisruptionBudget(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Deployments) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No deployments found"
		return result
	}

	// Build set of PDB selectors
	pdbSelectors := make(map[string]bool)
	for i := range resources.PodDisruptionBudgets {
		pdb := &resources.PodDisruptionBudgets[i]
		for k, v := range pdb.Spec.Selector.MatchLabels {
			pdbSelectors[k+"="+v] = true
		}
	}

	var count int
	for i := range resources.Deployments {
		deploy := &resources.Deployments[i]
		replicas := int32(1)
		if deploy.Spec.Replicas != nil {
			replicas = *deploy.Spec.Replicas
		}
		// Only check HA workloads (replicas > 1)
		if replicas < 2 {
			continue
		}

		// Check if any PDB selector matches the deployment's pod template labels
		matched := false
		for k, v := range deploy.Spec.Template.Labels {
			if pdbSelectors[k+"="+v] {
				matched = true
				break
			}
		}
		if !matched {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Deployment", Name: deploy.Name, Namespace: deploy.Namespace,
				Compliant: false,
				Message:   fmt.Sprintf("HA Deployment (%d replicas) has no PodDisruptionBudget", replicas),
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d HA deployment(s) missing PodDisruptionBudget", count)
	}
	return result
}
