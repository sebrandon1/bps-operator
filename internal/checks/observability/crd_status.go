package observability

import (
	"fmt"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// CheckCRDStatus verifies CRDs have a .status subresource defined.
// Checks both the subresource declaration and the schema properties.
func CheckCRDStatus(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.CRDs) == 0 {
		return result
	}

	var count int
	for i := range resources.CRDs {
		crd := &resources.CRDs[i]
		if !crdHasStatusSubresource(crd) {
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

func crdHasStatusSubresource(crd *apiextv1.CustomResourceDefinition) bool {
	for _, version := range crd.Spec.Versions {
		// Check subresource declaration
		if version.Subresources != nil && version.Subresources.Status != nil {
			return true
		}
		// Also check schema for status property
		if version.Schema != nil && version.Schema.OpenAPIV3Schema != nil {
			if _, ok := version.Schema.OpenAPIV3Schema.Properties["status"]; ok {
				return true
			}
		}
	}
	return false
}
