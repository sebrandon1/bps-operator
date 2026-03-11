package accesscontrol

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

func checkForbiddenCapability(resources *checks.DiscoveredResources, capName string) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var nonCompliant int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		allContainers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
		for j := range allContainers {
			container := &allContainers[j]
			if containerHasCapability(container, capName) {
				nonCompliant++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind:      "Pod",
					Name:      pod.Name,
					Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Container %q has %s capability", container.Name, capName),
				})
			}
		}
	}

	if nonCompliant > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) have %s capability", nonCompliant, capName)
	}
	return result
}

func containerHasCapability(container *corev1.Container, capName string) bool {
	if container.SecurityContext == nil || container.SecurityContext.Capabilities == nil {
		return false
	}
	for _, cap := range container.SecurityContext.Capabilities.Add {
		if strings.EqualFold(string(cap), "ALL") || strings.EqualFold(string(cap), capName) {
			return true
		}
	}
	return false
}

// CheckSysAdmin checks for SYS_ADMIN capability.
func CheckSysAdmin(resources *checks.DiscoveredResources) checks.CheckResult {
	return checkForbiddenCapability(resources, "SYS_ADMIN")
}

// CheckNetAdmin checks for NET_ADMIN capability.
func CheckNetAdmin(resources *checks.DiscoveredResources) checks.CheckResult {
	return checkForbiddenCapability(resources, "NET_ADMIN")
}

// CheckNetRaw checks for NET_RAW capability.
func CheckNetRaw(resources *checks.DiscoveredResources) checks.CheckResult {
	return checkForbiddenCapability(resources, "NET_RAW")
}

// CheckIPCLock checks for IPC_LOCK capability.
func CheckIPCLock(resources *checks.DiscoveredResources) checks.CheckResult {
	return checkForbiddenCapability(resources, "IPC_LOCK")
}

// CheckBPF checks for BPF capability.
func CheckBPF(resources *checks.DiscoveredResources) checks.CheckResult {
	return checkForbiddenCapability(resources, "BPF")
}
