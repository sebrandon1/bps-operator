package scanner

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// Discover lists all relevant resources in the target namespace.
func Discover(ctx context.Context, c client.Client, namespace string, labelSelector *metav1.LabelSelector) (*checks.DiscoveredResources, error) {
	resources := &checks.DiscoveredResources{
		Namespaces: []string{namespace},
	}

	listOpts := []client.ListOption{client.InNamespace(namespace)}
	if labelSelector != nil {
		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return nil, err
		}
		listOpts = append(listOpts, client.MatchingLabelsSelector{Selector: selector})
	}

	var pods corev1.PodList
	if err := c.List(ctx, &pods, listOpts...); err != nil {
		return nil, err
	}
	resources.Pods = pods.Items

	// Services, ServiceAccounts, RoleBindings use namespace only (no label filter)
	nsOpts := []client.ListOption{client.InNamespace(namespace)}

	var services corev1.ServiceList
	if err := c.List(ctx, &services, nsOpts...); err != nil {
		return nil, err
	}
	resources.Services = services.Items

	var serviceAccounts corev1.ServiceAccountList
	if err := c.List(ctx, &serviceAccounts, nsOpts...); err != nil {
		return nil, err
	}
	resources.ServiceAccounts = serviceAccounts.Items

	var roleBindings rbacv1.RoleBindingList
	if err := c.List(ctx, &roleBindings, nsOpts...); err != nil {
		return nil, err
	}
	resources.RoleBindings = roleBindings.Items

	// ClusterRoleBindings are cluster-scoped
	var crbs rbacv1.ClusterRoleBindingList
	if err := c.List(ctx, &crbs); err != nil {
		return nil, err
	}
	resources.ClusterRoleBindings = crbs.Items

	// CRDs are cluster-scoped
	var crds apiextv1.CustomResourceDefinitionList
	if err := c.List(ctx, &crds); err != nil {
		return nil, err
	}
	resources.CRDs = crds.Items

	// Deployments (namespace-scoped)
	var deployments appsv1.DeploymentList
	if err := c.List(ctx, &deployments, nsOpts...); err != nil {
		return nil, err
	}
	resources.Deployments = deployments.Items

	// StatefulSets (namespace-scoped)
	var statefulSets appsv1.StatefulSetList
	if err := c.List(ctx, &statefulSets, nsOpts...); err != nil {
		return nil, err
	}
	resources.StatefulSets = statefulSets.Items

	// DaemonSets (namespace-scoped)
	var daemonSets appsv1.DaemonSetList
	if err := c.List(ctx, &daemonSets, nsOpts...); err != nil {
		return nil, err
	}
	resources.DaemonSets = daemonSets.Items

	// NetworkPolicies (namespace-scoped)
	var netPolicies networkingv1.NetworkPolicyList
	if err := c.List(ctx, &netPolicies, nsOpts...); err != nil {
		return nil, err
	}
	resources.NetworkPolicies = netPolicies.Items

	// ResourceQuotas (namespace-scoped)
	var quotas corev1.ResourceQuotaList
	if err := c.List(ctx, &quotas, nsOpts...); err != nil {
		return nil, err
	}
	resources.ResourceQuotas = quotas.Items

	// PodDisruptionBudgets (namespace-scoped)
	var pdbs policyv1.PodDisruptionBudgetList
	if err := c.List(ctx, &pdbs, nsOpts...); err != nil {
		return nil, err
	}
	resources.PodDisruptionBudgets = pdbs.Items

	// Nodes (cluster-scoped)
	var nodes corev1.NodeList
	if err := c.List(ctx, &nodes); err != nil {
		return nil, err
	}
	resources.Nodes = nodes.Items

	// PersistentVolumes (cluster-scoped)
	var pvs corev1.PersistentVolumeList
	if err := c.List(ctx, &pvs); err != nil {
		return nil, err
	}
	resources.PersistentVolumes = pvs.Items

	// StorageClasses (cluster-scoped)
	var scs storagev1.StorageClassList
	if err := c.List(ctx, &scs); err != nil {
		return nil, err
	}
	resources.StorageClasses = scs.Items

	return resources, nil
}
