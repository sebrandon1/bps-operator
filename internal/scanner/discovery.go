package scanner

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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

	return resources, nil
}
