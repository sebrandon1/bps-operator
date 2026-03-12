package manageability

import "github.com/sebrandon1/bps-operator/internal/checks"

func init() {
	checks.Register(checks.CheckInfo{
		Name: "manageability-container-port-name-format", Category: "manageability",
		Description: "Verifies container port names follow IANA naming conventions",
		Remediation: "Use IANA-compliant port names (lowercase, alphanumeric, hyphens, max 15 chars)",
		CatalogID:   "manageability-container-port-name-format",
		Fn:          CheckPortNameFormat,
	})
	checks.Register(checks.CheckInfo{
		Name: "manageability-containers-image-tag", Category: "manageability",
		Description: "Verifies container images use a digest or specific tag (not :latest)",
		Remediation: "Use a specific image tag or digest reference instead of :latest",
		CatalogID:   "manageability-containers-image-tag",
		Fn:          CheckImageTag,
	})
}
