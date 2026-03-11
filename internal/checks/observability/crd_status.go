package observability

import (
	"fmt"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// CheckCRDStatus verifies CRDs have a .status subresource defined.
func CheckCRDStatus(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.CRDs) == 0 {
		return result
	}

	var count int
	for i := range resources.CRDs {
		crd := &resources.CRDs[i]
		hasStatus := false
		for _, version := range crd.Spec.Versions {
			if version.Subresources != nil && version.Subresources.Status != nil {
				hasStatus = true
				break
			}
		}
		if !hasStatus {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "CustomResourceDefinition", Name: crd.Name, Namespace: "",
				Compliant: false,
				Message:   fmt.Sprintf("CRD %q does not define a .status subresource", crd.Name),
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d CRD(s) missing .status subresource", count)
	}
	return result
}
