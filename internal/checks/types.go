package checks

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// DiscoveredResources holds all resources discovered in the target namespace.
type DiscoveredResources struct {
	Pods                []corev1.Pod
	Services            []corev1.Service
	ServiceAccounts     []corev1.ServiceAccount
	RoleBindings        []rbacv1.RoleBinding
	ClusterRoleBindings []rbacv1.ClusterRoleBinding
	CRDs                []apiextv1.CustomResourceDefinition
	Namespaces          []string
	ProbePods           map[string]*corev1.Pod // node name → probe pod
	ProbeExecutor       ProbeExecutor

	// Phase 2A additions
	Deployments  []appsv1.Deployment
	StatefulSets []appsv1.StatefulSet
	DaemonSets   []appsv1.DaemonSet

	// Phase 2B additions
	NetworkPolicies []networkingv1.NetworkPolicy

	// Phase 2C additions (OLM types not added yet — requires operator-framework/api dependency)

	// Phase 2D additions
	ResourceQuotas      []corev1.ResourceQuota
	Nodes               []corev1.Node
	PersistentVolumes   []corev1.PersistentVolume
	StorageClasses      []storagev1.StorageClass
	PodDisruptionBudgets []policyv1.PodDisruptionBudget
}

// ProbeExecutor allows checks to exec commands on probe pods.
type ProbeExecutor interface {
	ExecCommand(ctx context.Context, pod *corev1.Pod, command string) (stdout, stderr string, err error)
}

// CheckFunc is the signature for a best practice check function.
type CheckFunc func(resources *DiscoveredResources) CheckResult

// CheckResult holds the outcome of a single check.
type CheckResult struct {
	ComplianceStatus string
	Reason           string
	Details          []ResourceDetail
}

// ResourceDetail describes a specific resource's compliance status.
type ResourceDetail struct {
	Kind      string
	Name      string
	Namespace string
	Compliant bool
	Message   string
}

// CheckInfo describes a registered check.
type CheckInfo struct {
	Name        string
	Category    string
	Description string
	Remediation string
	CatalogID   string // Anchor in certsuite CATALOG.md (e.g. "access-control-no-1337-uid")
	Fn          CheckFunc
}
