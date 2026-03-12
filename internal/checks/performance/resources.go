package performance

import (
	"fmt"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// CheckExclusiveCPUPool verifies containers requesting whole CPUs have Guaranteed QoS.
// A container should have CPU requests == limits with whole-number values for exclusive CPU pinning.
func CheckExclusiveCPUPool(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		for j := range pod.Spec.Containers {
			container := &pod.Spec.Containers[j]
			cpuReq := container.Resources.Requests.Cpu()
			cpuLim := container.Resources.Limits.Cpu()

			// Only check containers that request whole CPUs (>= 1000m and integer values)
			if cpuReq.IsZero() || cpuReq.MilliValue()%1000 != 0 {
				continue
			}

			// For whole-CPU requests, limits must equal requests
			if cpuLim.IsZero() || !cpuReq.Equal(*cpuLim) {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Container %q requests %s whole CPUs but limits (%s) do not match", container.Name, cpuReq.String(), cpuLim.String()),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) have mismatched exclusive CPU pool configuration", count)
	}
	return result
}

// CheckRTAppsNoExecProbes verifies RT (real-time) containers don't use exec probes.
// RT containers are identified by having the "rt-app" or "realtime" annotation,
// or by requesting whole CPUs (indicating CPU pinning for RT workloads).
func CheckRTAppsNoExecProbes(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		// Check if pod is marked as real-time
		isRT := pod.Annotations["rt-app"] == "true" || pod.Annotations["realtime"] == "true"
		if !isRT {
			continue
		}

		for j := range pod.Spec.Containers {
			container := &pod.Spec.Containers[j]
			hasExecProbe := false
			if container.LivenessProbe != nil && container.LivenessProbe.Exec != nil {
				hasExecProbe = true
			}
			if container.ReadinessProbe != nil && container.ReadinessProbe.Exec != nil {
				hasExecProbe = true
			}
			if container.StartupProbe != nil && container.StartupProbe.Exec != nil {
				hasExecProbe = true
			}
			if hasExecProbe {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("RT container %q uses exec probe", container.Name),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d RT container(s) use exec probes", count)
	}
	return result
}

// CheckMemoryLimit verifies containers have memory limits set.
func CheckMemoryLimit(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		for j := range pod.Spec.Containers {
			container := &pod.Spec.Containers[j]
			memLim := container.Resources.Limits.Memory()
			if memLim.IsZero() {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Container %q does not have memory limits set", container.Name),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) missing memory limits", count)
	}
	return result
}
