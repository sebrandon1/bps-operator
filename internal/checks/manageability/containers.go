package manageability

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// ianaPortNameRegex matches IANA service name format:
// - 1-15 characters
// - lowercase alphanumeric and hyphens
// - must begin and end with alphanumeric
// - must contain at least one letter
var ianaPortNameRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,13}[a-z0-9])?$`)

// CheckPortNameFormat verifies container port names follow IANA naming conventions.
func CheckPortNameFormat(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		for _, container := range pod.Spec.Containers {
			for _, port := range container.Ports {
				if port.Name == "" {
					continue // unnamed ports are not checked
				}
				if !ianaPortNameRegex.MatchString(port.Name) {
					count++
					result.Details = append(result.Details, checks.ResourceDetail{
						Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
						Compliant: false,
						Message:   fmt.Sprintf("Container %q port name %q does not follow IANA format", container.Name, port.Name),
					})
				}
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d port name(s) do not follow IANA format", count)
	}
	return result
}

// CheckImageTag verifies container images use a digest or specific tag, not :latest.
func CheckImageTag(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		allContainers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
		for j := range allContainers {
			container := &allContainers[j]
			if isLatestOrUntagged(container.Image) {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Container %q uses image %q (latest or untagged)", container.Name, container.Image),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) use :latest or untagged images", count)
	}
	return result
}

// isLatestOrUntagged returns true if the image reference uses :latest or has no tag/digest.
func isLatestOrUntagged(image string) bool {
	// If it has a digest, it's fine
	if strings.Contains(image, "@sha256:") {
		return false
	}
	// Check for explicit :latest
	if strings.HasSuffix(image, ":latest") {
		return true
	}
	// Check for no tag at all (e.g. "nginx" or "registry.io/image")
	// After removing the registry prefix, if there's no ":", it's untagged
	parts := strings.Split(image, "/")
	lastPart := parts[len(parts)-1]
	return !strings.Contains(lastPart, ":")
}
