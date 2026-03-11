package accesscontrol

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// CheckNonRootUser verifies containers run as non-root.
// A container is compliant if RunAsNonRoot=true OR RunAsUser is set to a non-zero value.
// Matches certsuite GetRunAsNonRootFalseContainers logic.
func CheckNonRootUser(resources *checks.DiscoveredResources) checks.CheckResult {
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
			if !isContainerRunAsNonRoot(pod, container) && !isContainerRunAsNonRootUserID(pod, container) {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Container %q does not have runAsNonRoot=true or runAsUser!=0", container.Name),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) may run as root", count)
	}
	return result
}

// isContainerRunAsNonRoot checks if RunAsNonRoot is explicitly true at container or pod level.
func isContainerRunAsNonRoot(pod *corev1.Pod, container *corev1.Container) bool {
	if container.SecurityContext != nil && container.SecurityContext.RunAsNonRoot != nil {
		return *container.SecurityContext.RunAsNonRoot
	}
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.RunAsNonRoot != nil {
		return *pod.Spec.SecurityContext.RunAsNonRoot
	}
	return false
}

// isContainerRunAsNonRootUserID checks if RunAsUser is set to a non-zero value.
func isContainerRunAsNonRootUserID(pod *corev1.Pod, container *corev1.Container) bool {
	if container.SecurityContext != nil && container.SecurityContext.RunAsUser != nil {
		return *container.SecurityContext.RunAsUser != 0
	}
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.RunAsUser != nil {
		return *pod.Spec.SecurityContext.RunAsUser != 0
	}
	return false
}

// CheckPrivilegeEscalation verifies containers do not allow privilege escalation.
// Matches certsuite: only flags when AllowPrivilegeEscalation is explicitly true.
// Nil (unset) is treated as compliant.
func CheckPrivilegeEscalation(resources *checks.DiscoveredResources) checks.CheckResult {
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
			if container.SecurityContext != nil &&
				container.SecurityContext.AllowPrivilegeEscalation != nil &&
				*container.SecurityContext.AllowPrivilegeEscalation {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Container %q has allowPrivilegeEscalation set to true", container.Name),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) allow privilege escalation", count)
	}
	return result
}

// CheckReadOnlyFilesystem verifies containers set readOnlyRootFilesystem to true.
// Matches certsuite: nil or false is non-compliant.
func CheckReadOnlyFilesystem(resources *checks.DiscoveredResources) checks.CheckResult {
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
			if container.SecurityContext == nil ||
				container.SecurityContext.ReadOnlyRootFilesystem == nil ||
				!*container.SecurityContext.ReadOnlyRootFilesystem {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Container %q does not set readOnlyRootFilesystem to true", container.Name),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) do not set readOnlyRootFilesystem", count)
	}
	return result
}

// Check1337UID verifies pods do not run as UID 1337 (Istio conflict).
// Matches certsuite: only checks pod-level SecurityContext.RunAsUser.
func Check1337UID(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		if pod.Spec.SecurityContext != nil &&
			pod.Spec.SecurityContext.RunAsUser != nil &&
			*pod.Spec.SecurityContext.RunAsUser == 1337 {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
				Compliant: false,
				Message:   "Pod SecurityContext RunAsUser is set to 1337 (reserved by Istio)",
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d pod(s) use UID 1337", count)
	}
	return result
}

// CheckSecurityContext categorizes container security contexts (SCC check).
// Matches certsuite: non-compliant if category > CategoryID1NoUID0.
// Simplified: flags containers that are privileged, have host-level caps, or run as root with escalation.
func CheckSecurityContext(resources *checks.DiscoveredResources) checks.CheckResult {
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
			category := categorizeSCC(pod, container)
			if category > categoryID1NoUID0 {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Container %q requires elevated SCC (category %d)", container.Name, category),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) require elevated SCC", count)
	}
	return result
}

// SCC category constants matching certsuite securitycontextcontainer categories.
const (
	categoryID1        = 0 // Most restrictive
	categoryID1NoUID0  = 1 // Restrictive but no UID 0 requirement
	categoryID2        = 2 // Needs some elevated privs
	categoryID3        = 3 // Needs significant elevated privs
	categoryID4        = 4 // Privileged
)

// categorizeSCC assigns an SCC category to a container.
// Categories match certsuite securitycontextcontainer.GetContainerSCC:
//   - CategoryID1: fully restricted
//   - CategoryID1NoUID0: restricted, no UID 0 enforcement
//   - CategoryID2+: needs elevated privileges
func categorizeSCC(pod *corev1.Pod, container *corev1.Container) int {
	if container.SecurityContext == nil {
		return categoryID1NoUID0
	}
	sc := container.SecurityContext

	// Category 4: Privileged
	if sc.Privileged != nil && *sc.Privileged {
		return categoryID4
	}

	// Category 3: Has host-level ports
	for _, port := range container.Ports {
		if port.HostPort != 0 {
			return categoryID3
		}
	}

	// Category 3: Has dangerous capabilities
	if sc.Capabilities != nil {
		for _, cap := range sc.Capabilities.Add {
			switch string(cap) {
			case "ALL", "SYS_ADMIN", "NET_ADMIN", "NET_RAW", "IPC_LOCK", "BPF":
				return categoryID3
			}
		}
	}

	// Category 2: AllowPrivilegeEscalation=true
	if sc.AllowPrivilegeEscalation != nil && *sc.AllowPrivilegeEscalation {
		return categoryID2
	}

	// Category 1NoUID0: RunAsNonRoot not enforced
	if sc.RunAsNonRoot == nil || !*sc.RunAsNonRoot {
		// Check pod level too
		podNonRoot := pod.Spec.SecurityContext != nil &&
			pod.Spec.SecurityContext.RunAsNonRoot != nil &&
			*pod.Spec.SecurityContext.RunAsNonRoot
		if !podNonRoot {
			return categoryID1NoUID0
		}
	}

	return categoryID1
}
