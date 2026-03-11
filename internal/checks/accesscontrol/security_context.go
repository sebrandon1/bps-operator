package accesscontrol

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// CheckNonRootUser verifies containers have runAsNonRoot set to true.
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
			if !isRunAsNonRoot(pod, container) {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Container %q does not have runAsNonRoot set to true", container.Name),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) do not set runAsNonRoot", count)
	}
	return result
}

func isRunAsNonRoot(pod *corev1.Pod, container *corev1.Container) bool {
	// Container-level takes precedence
	if container.SecurityContext != nil && container.SecurityContext.RunAsNonRoot != nil {
		return *container.SecurityContext.RunAsNonRoot
	}
	// Fall back to pod-level
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.RunAsNonRoot != nil {
		return *pod.Spec.SecurityContext.RunAsNonRoot
	}
	return false
}

// CheckPrivilegeEscalation verifies containers set allowPrivilegeEscalation to false.
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
			if container.SecurityContext == nil ||
				container.SecurityContext.AllowPrivilegeEscalation == nil ||
				*container.SecurityContext.AllowPrivilegeEscalation {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Container %q does not set allowPrivilegeEscalation to false", container.Name),
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

// Check1337UID verifies containers do not run as UID 1337 (Istio conflict).
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
		allContainers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
		for j := range allContainers {
			container := &allContainers[j]
			uid := getRunAsUser(pod, container)
			if uid != nil && *uid == 1337 {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Container %q runs as UID 1337 (reserved by Istio)", container.Name),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) use UID 1337", count)
	}
	return result
}

func getRunAsUser(pod *corev1.Pod, container *corev1.Container) *int64 {
	if container.SecurityContext != nil && container.SecurityContext.RunAsUser != nil {
		return container.SecurityContext.RunAsUser
	}
	if pod.Spec.SecurityContext != nil {
		return pod.Spec.SecurityContext.RunAsUser
	}
	return nil
}

// CheckSecurityContext categorizes container security contexts (SCC check).
// A container is non-compliant if it runs as privileged.
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
			category := categorizeSCC(container)
			if category == "privileged" {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Container %q is categorized as %q SCC", container.Name, category),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) require privileged SCC", count)
	}
	return result
}

func categorizeSCC(container *corev1.Container) string {
	if container.SecurityContext == nil {
		return "restricted"
	}
	sc := container.SecurityContext
	if sc.Privileged != nil && *sc.Privileged {
		return "privileged"
	}
	if sc.AllowPrivilegeEscalation != nil && *sc.AllowPrivilegeEscalation {
		return "anyuid"
	}
	if sc.RunAsNonRoot == nil || !*sc.RunAsNonRoot {
		return "anyuid"
	}
	return "restricted"
}
