package lifecycle

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// CheckImagePullPolicy verifies imagePullPolicy is Always or image uses a digest.
func CheckImagePullPolicy(resources *checks.DiscoveredResources) checks.CheckResult {
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
			if container.ImagePullPolicy == corev1.PullAlways {
				continue
			}
			// IfNotPresent is compliant if image uses a digest
			if container.ImagePullPolicy == corev1.PullIfNotPresent && strings.Contains(container.Image, "@sha256:") {
				continue
			}
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
				Compliant: false,
				Message:   fmt.Sprintf("Container %q has imagePullPolicy %q without digest reference", container.Name, container.ImagePullPolicy),
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) have non-compliant imagePullPolicy", count)
	}
	return result
}

// allowedOwnerKinds are the workload controller kinds that should own pods.
var allowedOwnerKinds = map[string]bool{
	"ReplicaSet":  true,
	"StatefulSet": true,
	"DaemonSet":   true,
	"Job":         true,
}

// CheckPodOwnerType verifies pods are owned by a workload controller.
func CheckPodOwnerType(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		if !hasAllowedOwner(pod) {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
				Compliant: false,
				Message:   "Pod is not owned by ReplicaSet, StatefulSet, DaemonSet, or Job",
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d pod(s) are not managed by a workload controller", count)
	}
	return result
}

func hasAllowedOwner(pod *corev1.Pod) bool {
	for _, ref := range pod.OwnerReferences {
		if allowedOwnerKinds[ref.Kind] {
			return true
		}
	}
	return false
}

// CheckPodScheduling verifies pods have scheduling directives.
func CheckPodScheduling(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		hasNodeSelector := len(pod.Spec.NodeSelector) > 0
		hasAffinity := pod.Spec.Affinity != nil
		hasTolerations := len(pod.Spec.Tolerations) > 0
		if !hasNodeSelector && !hasAffinity && !hasTolerations {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
				Compliant: false,
				Message:   "Pod has no nodeSelector, affinity, or tolerations",
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d pod(s) have no scheduling directives", count)
	}
	return result
}

// CheckHighAvailability verifies Deployments have replicas > 1.
func CheckHighAvailability(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Deployments) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No deployments found"
		return result
	}

	var count int
	for i := range resources.Deployments {
		deploy := &resources.Deployments[i]
		replicas := int32(1)
		if deploy.Spec.Replicas != nil {
			replicas = *deploy.Spec.Replicas
		}
		if replicas < 2 {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Deployment", Name: deploy.Name, Namespace: deploy.Namespace,
				Compliant: false,
				Message:   fmt.Sprintf("Deployment has %d replica(s), expected at least 2", replicas),
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d deployment(s) have fewer than 2 replicas", count)
	}
	return result
}

// CheckCPUIsolation verifies CPU requests equal CPU limits (Guaranteed QoS for CPU).
func CheckCPUIsolation(resources *checks.DiscoveredResources) checks.CheckResult {
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
			if cpuReq.IsZero() || cpuLim.IsZero() || !cpuReq.Equal(*cpuLim) {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Container %q CPU requests (%s) != limits (%s)", container.Name, cpuReq.String(), cpuLim.String()),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) do not have CPU requests equal to limits", count)
	}
	return result
}

// CheckAffinityRequired verifies pods have podAntiAffinity for high availability.
func CheckAffinityRequired(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		if pod.Spec.Affinity == nil || pod.Spec.Affinity.PodAntiAffinity == nil {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
				Compliant: false,
				Message:   "Pod does not have podAntiAffinity configured",
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d pod(s) missing podAntiAffinity", count)
	}
	return result
}

// masterTaintKeys are taint keys used by control-plane/master nodes.
var masterTaintKeys = map[string]bool{
	"node-role.kubernetes.io/master":        true,
	"node-role.kubernetes.io/control-plane": true,
}

// CheckTolerationBypass verifies pods do not tolerate master/control-plane taints.
func CheckTolerationBypass(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		for _, tol := range pod.Spec.Tolerations {
			if masterTaintKeys[tol.Key] && (tol.Effect == corev1.TaintEffectNoSchedule || tol.Effect == corev1.TaintEffectNoExecute) {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Pod tolerates master taint %q with effect %s", tol.Key, tol.Effect),
				})
				break
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d pod(s) tolerate master/control-plane taints", count)
	}
	return result
}

// CheckPVReclaimPolicy verifies PersistentVolume reclaimPolicy is not Delete.
func CheckPVReclaimPolicy(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.PersistentVolumes) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No PersistentVolumes found"
		return result
	}

	var count int
	for i := range resources.PersistentVolumes {
		pv := &resources.PersistentVolumes[i]
		if pv.Spec.PersistentVolumeReclaimPolicy == corev1.PersistentVolumeReclaimDelete {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "PersistentVolume", Name: pv.Name,
				Compliant: false,
				Message:   "PersistentVolume reclaimPolicy is Delete",
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d PersistentVolume(s) have reclaimPolicy Delete", count)
	}
	return result
}

// CheckStorageProvisioner verifies StorageClasses have a non-empty provisioner.
func CheckStorageProvisioner(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.StorageClasses) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No StorageClasses found"
		return result
	}

	var count int
	for i := range resources.StorageClasses {
		sc := &resources.StorageClasses[i]
		if sc.Provisioner == "" || sc.Provisioner == "kubernetes.io/no-provisioner" {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "StorageClass", Name: sc.Name,
				Compliant: false,
				Message:   fmt.Sprintf("StorageClass provisioner is %q", sc.Provisioner),
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d StorageClass(es) have invalid provisioner", count)
	}
	return result
}
