package accesscontrol

import (
	"fmt"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// defaultNamespaces are Kubernetes/OpenShift system namespaces where workloads should not run.
var defaultNamespaces = map[string]bool{
	"default":                          true,
	"kube-system":                      true,
	"kube-public":                      true,
	"kube-node-lease":                  true,
	"openshift":                        true,
	"openshift-apiserver":              true,
	"openshift-authentication":         true,
	"openshift-config":                 true,
	"openshift-console":                true,
	"openshift-controller-manager":     true,
	"openshift-dns":                    true,
	"openshift-etcd":                   true,
	"openshift-image-registry":         true,
	"openshift-infra":                  true,
	"openshift-ingress":                true,
	"openshift-kube-apiserver":         true,
	"openshift-kube-controller-manager": true,
	"openshift-kube-scheduler":         true,
	"openshift-machine-api":            true,
	"openshift-machine-config-operator": true,
	"openshift-marketplace":            true,
	"openshift-monitoring":             true,
	"openshift-multus":                 true,
	"openshift-network-operator":       true,
	"openshift-node":                   true,
	"openshift-operator-lifecycle-manager": true,
	"openshift-operators":              true,
	"openshift-sdn":                    true,
}

// CheckNamespace verifies pods run in allowed namespaces.
func CheckNamespace(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		if defaultNamespaces[pod.Namespace] {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
				Compliant: false,
				Message:   fmt.Sprintf("Pod is running in system namespace %q", pod.Namespace),
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d pod(s) are running in system namespaces", count)
	}
	return result
}

// CheckNamespaceResourceQuota verifies the namespace has a ResourceQuota defined.
func CheckNamespaceResourceQuota(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if len(resources.Namespaces) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No namespaces found"
		return result
	}

	if len(resources.ResourceQuotas) == 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = "No ResourceQuota found in namespace"
		if len(resources.Namespaces) > 0 {
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Namespace", Name: resources.Namespaces[0],
				Compliant: false,
				Message:   "Namespace does not have a ResourceQuota defined",
			})
		}
	}
	return result
}
