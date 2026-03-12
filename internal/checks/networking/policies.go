package networking

import (
	networkingv1 "k8s.io/api/networking/v1"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// CheckNetworkPolicyDenyAll verifies a default-deny NetworkPolicy exists.
// A default-deny policy selects all pods ({}) and has empty ingress/egress rules.
func CheckNetworkPolicyDenyAll(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	hasDenyIngress := false
	hasDenyEgress := false

	for i := range resources.NetworkPolicies {
		np := &resources.NetworkPolicies[i]

		// A default-deny policy has an empty pod selector (selects all pods)
		if len(np.Spec.PodSelector.MatchLabels) > 0 || len(np.Spec.PodSelector.MatchExpressions) > 0 {
			continue
		}

		for _, pt := range np.Spec.PolicyTypes {
			if pt == networkingv1.PolicyTypeIngress && len(np.Spec.Ingress) == 0 {
				hasDenyIngress = true
			}
			if pt == networkingv1.PolicyTypeEgress && len(np.Spec.Egress) == 0 {
				hasDenyEgress = true
			}
		}
	}

	if !hasDenyIngress || !hasDenyEgress {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = "No default-deny NetworkPolicy found for both ingress and egress"
		if len(resources.Namespaces) > 0 {
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Namespace", Name: resources.Namespaces[0],
				Compliant: false,
				Message:   "Namespace is missing a default-deny NetworkPolicy",
			})
		}
	}
	return result
}
