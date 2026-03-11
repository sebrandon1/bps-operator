package accesscontrol

import (
	"fmt"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// CheckHostNetwork verifies pods do not use HostNetwork.
func CheckHostNetwork(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		if pod.Spec.HostNetwork {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
				Compliant: false, Message: "HostNetwork is set to true",
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d pod(s) have HostNetwork enabled", count)
	}
	return result
}

// CheckHostPath verifies pods do not use HostPath volumes.
func CheckHostPath(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		for _, vol := range pod.Spec.Volumes {
			if vol.HostPath != nil && vol.HostPath.Path != "" {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Volume %q uses HostPath %s", vol.Name, vol.HostPath.Path),
				})
				break // one detail per pod
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d pod(s) use HostPath volumes", count)
	}
	return result
}

// CheckHostIPC verifies pods do not use HostIPC.
func CheckHostIPC(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		if pod.Spec.HostIPC {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
				Compliant: false, Message: "HostIPC is set to true",
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d pod(s) have HostIPC enabled", count)
	}
	return result
}

// CheckHostPID verifies pods do not use HostPID.
func CheckHostPID(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		if pod.Spec.HostPID {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
				Compliant: false, Message: "HostPID is set to true",
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d pod(s) have HostPID enabled", count)
	}
	return result
}

// CheckContainerHostPort verifies containers do not use HostPort.
func CheckContainerHostPort(resources *checks.DiscoveredResources) checks.CheckResult {
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
				if port.HostPort != 0 {
					count++
					result.Details = append(result.Details, checks.ResourceDetail{
						Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
						Compliant: false,
						Message:   fmt.Sprintf("Container %q uses HostPort %d", container.Name, port.HostPort),
					})
				}
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) use HostPort", count)
	}
	return result
}
