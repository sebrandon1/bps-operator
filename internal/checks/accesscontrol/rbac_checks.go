package accesscontrol

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// CheckServiceAccount verifies pods do not use the default service account.
func CheckServiceAccount(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		sa := pod.Spec.ServiceAccountName
		if sa == "" || sa == "default" {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
				Compliant: false, Message: "Pod uses the default service account",
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d pod(s) use the default service account", count)
	}
	return result
}

// CheckRoleBindings verifies RoleBindings do not reference service accounts from outside target namespaces.
func CheckRoleBindings(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.RoleBindings) == 0 {
		return result
	}

	targetNS := make(map[string]bool, len(resources.Namespaces))
	for _, ns := range resources.Namespaces {
		targetNS[ns] = true
	}

	podSAs := make(map[string]bool)
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		sa := pod.Spec.ServiceAccountName
		if sa == "" {
			sa = "default"
		}
		podSAs[pod.Namespace+"/"+sa] = true
	}

	var count int
	for i := range resources.RoleBindings {
		rb := &resources.RoleBindings[i]
		for _, subject := range rb.Subjects {
			if subject.Kind != "ServiceAccount" {
				continue
			}
			subjectNS := subject.Namespace
			if subjectNS == "" {
				subjectNS = rb.Namespace
			}
			saKey := subjectNS + "/" + subject.Name
			if !podSAs[saKey] {
				continue
			}
			if !targetNS[subjectNS] {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "RoleBinding", Name: rb.Name, Namespace: rb.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("References ServiceAccount %s/%s from non-target namespace", subjectNS, subject.Name),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d RoleBinding(s) reference non-target namespace ServiceAccounts", count)
	}
	return result
}

// CheckClusterRoleBindings verifies pods are not linked to ClusterRoleBindings.
func CheckClusterRoleBindings(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.ClusterRoleBindings) == 0 || len(resources.Pods) == 0 {
		return result
	}

	podSAs := make(map[string]bool)
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		sa := pod.Spec.ServiceAccountName
		if sa == "" {
			sa = "default"
		}
		podSAs[pod.Namespace+"/"+sa] = true
	}

	var count int
	for i := range resources.ClusterRoleBindings {
		crb := &resources.ClusterRoleBindings[i]
		for _, subject := range crb.Subjects {
			if subject.Kind != "ServiceAccount" {
				continue
			}
			saKey := subject.Namespace + "/" + subject.Name
			if podSAs[saKey] {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "ClusterRoleBinding", Name: crb.Name, Namespace: "",
					Compliant: false,
					Message:   fmt.Sprintf("Binds ServiceAccount %s used by pod(s)", saKey),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d ClusterRoleBinding(s) bind pod ServiceAccounts", count)
	}
	return result
}

// CheckAutomountToken verifies pods do not automount service account tokens.
func CheckAutomountToken(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	saAutomount := make(map[string]*bool)
	for i := range resources.ServiceAccounts {
		sa := &resources.ServiceAccounts[i]
		saAutomount[sa.Namespace+"/"+sa.Name] = sa.AutomountServiceAccountToken
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		if automountEnabled(pod, saAutomount) {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
				Compliant: false, Message: "Service account token is automounted",
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d pod(s) automount service account tokens", count)
	}
	return result
}

// automountEnabled checks if the service account token is automounted for a pod.
// Pod-level setting takes precedence over SA-level setting. Default is true (mounted).
func automountEnabled(pod *corev1.Pod, saAutomount map[string]*bool) bool {
	// Pod-level setting takes precedence
	if pod.Spec.AutomountServiceAccountToken != nil {
		return *pod.Spec.AutomountServiceAccountToken
	}
	// Fall back to SA-level
	saName := pod.Spec.ServiceAccountName
	if saName == "" {
		saName = "default"
	}
	saKey := pod.Namespace + "/" + saName
	if saVal, ok := saAutomount[saKey]; ok && saVal != nil {
		return *saVal
	}
	// Default: automount is enabled
	return true
}
