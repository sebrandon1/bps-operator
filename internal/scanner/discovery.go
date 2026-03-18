package scanner

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	netattdefv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmpackagev1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	apiserverv1 "github.com/openshift/api/apiserver/v1"
	configv1 "github.com/openshift/api/config/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/redhat-best-practices-for-k8s/checks"
)

// Discover lists all relevant resources in the target namespace.
func Discover(ctx context.Context, c client.Client, namespace string, labelSelector *metav1.LabelSelector, discoveryClient discovery.ServerVersionInterface) (*checks.DiscoveredResources, error) {
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

	var roles rbacv1.RoleList
	if err := c.List(ctx, &roles, nsOpts...); err != nil {
		return nil, err
	}
	resources.Roles = roles.Items

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

	// Helm chart releases (secrets with type helm.sh/release.v1)
	var secrets corev1.SecretList
	if err := c.List(ctx, &secrets, nsOpts...); err != nil {
		return nil, err
	}
	for i := range secrets.Items {
		if secrets.Items[i].Type == "helm.sh/release.v1" {
			if release, ok := parseHelmRelease(&secrets.Items[i]); ok {
				resources.HelmChartReleases = append(resources.HelmChartReleases, release)
			}
		}
	}

	// K8s version
	if discoveryClient != nil {
		if info, err := discoveryClient.ServerVersion(); err == nil {
			resources.K8sVersion = info.GitVersion
		}
	}

	// Scalable CRD resources
	resources.ScalableResources = discoverScalableResources(ctx, c, crds.Items, namespace)

	// --- OpenShift-specific resources (graceful skip) ---

	// ClusterVersion (singleton named "version")
	var cv configv1.ClusterVersion
	if err := tryGet(ctx, c, types.NamespacedName{Name: "version"}, &cv); err == nil {
		resources.ClusterVersion = &cv
		resources.OpenshiftVersion = extractOpenshiftVersion(&cv)
		resources.OCPStatus = deriveOCPStatus(&cv)
	}

	// ClusterOperators
	var cos configv1.ClusterOperatorList
	if items, err := tryList(ctx, c, &cos); err == nil {
		resources.ClusterOperators = items.([]configv1.ClusterOperator)
	}

	// APIRequestCounts
	var arcs apiserverv1.APIRequestCountList
	if items, err := tryList(ctx, c, &arcs); err == nil {
		resources.APIRequestCounts = items.([]apiserverv1.APIRequestCount)
	}

	// --- OLM resources (graceful skip) ---

	// CSVs (namespace-scoped)
	var csvs olmv1alpha1.ClusterServiceVersionList
	if items, err := tryListNS(ctx, c, &csvs, namespace); err == nil {
		resources.CSVs = items.([]olmv1alpha1.ClusterServiceVersion)
	}

	// CatalogSources (cluster-wide)
	var catalogs olmv1alpha1.CatalogSourceList
	if items, err := tryList(ctx, c, &catalogs); err == nil {
		resources.CatalogSources = items.([]olmv1alpha1.CatalogSource)
	}

	// Subscriptions (namespace-scoped)
	var subs olmv1alpha1.SubscriptionList
	if items, err := tryListNS(ctx, c, &subs, namespace); err == nil {
		resources.Subscriptions = items.([]olmv1alpha1.Subscription)
	}

	// PackageManifests (namespace-scoped)
	var pkgs olmpackagev1.PackageManifestList
	if items, err := tryListNS(ctx, c, &pkgs, namespace); err == nil {
		resources.PackageManifests = items.([]olmpackagev1.PackageManifest)
	}

	// --- Networking resources (graceful skip) ---

	// NetworkAttachmentDefinitions
	var nads netattdefv1.NetworkAttachmentDefinitionList
	if items, err := tryListNS(ctx, c, &nads, namespace); err == nil {
		resources.NetworkAttachmentDefinitions = items.([]netattdefv1.NetworkAttachmentDefinition)
	}

	// SR-IOV resources (unstructured)
	resources.SriovNetworks = listUnstructured(ctx, c, schema.GroupVersionResource{
		Group: "sriovnetwork.openshift.io", Version: "v1", Resource: "sriovnetworks",
	}, namespace)
	resources.SriovNetworkNodePolicies = listUnstructured(ctx, c, schema.GroupVersionResource{
		Group: "sriovnetwork.openshift.io", Version: "v1", Resource: "sriovnetworknodepolicies",
	}, "")

	return resources, nil
}

// tryGet attempts to get a resource and gracefully returns an error if the API is not available.
func tryGet(ctx context.Context, c client.Client, key types.NamespacedName, obj client.Object) error {
	err := c.Get(ctx, key, obj)
	if err == nil {
		return nil
	}
	if errors.IsNotFound(err) || meta.IsNoMatchError(err) || errors.IsNotFound(err) {
		return err
	}
	return err
}

// tryList attempts to list resources. Returns the items slice and nil error on success.
// Returns nil, error when the CRD is not registered (graceful skip).
func tryList(ctx context.Context, c client.Client, list client.ObjectList) (any, error) {
	return doTryList(ctx, c, list)
}

// tryListNS attempts to list namespaced resources.
func tryListNS(ctx context.Context, c client.Client, list client.ObjectList, namespace string) (any, error) {
	return doTryList(ctx, c, list, client.InNamespace(namespace))
}

func doTryList(ctx context.Context, c client.Client, list client.ObjectList, opts ...client.ListOption) (any, error) {
	if err := c.List(ctx, list, opts...); err != nil {
		if meta.IsNoMatchError(err) || errors.IsNotFound(err) {
			return nil, err
		}
		return nil, err
	}
	// Extract items from the typed list
	switch l := list.(type) {
	case *configv1.ClusterOperatorList:
		return l.Items, nil
	case *apiserverv1.APIRequestCountList:
		return l.Items, nil
	case *olmv1alpha1.ClusterServiceVersionList:
		return l.Items, nil
	case *olmv1alpha1.CatalogSourceList:
		return l.Items, nil
	case *olmv1alpha1.SubscriptionList:
		return l.Items, nil
	case *olmpackagev1.PackageManifestList:
		return l.Items, nil
	case *netattdefv1.NetworkAttachmentDefinitionList:
		return l.Items, nil
	default:
		return nil, fmt.Errorf("unsupported list type %T", list)
	}
}

// listUnstructured lists unstructured resources, returning empty on any error (graceful skip).
func listUnstructured(ctx context.Context, c client.Client, gvr schema.GroupVersionResource, namespace string) []unstructured.Unstructured {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gvr.Group,
		Version: gvr.Version,
		Kind:    gvr.Resource, // Will be resolved by the client
	})

	opts := []client.ListOption{}
	if namespace != "" {
		opts = append(opts, client.InNamespace(namespace))
	}

	if err := c.List(ctx, list, opts...); err != nil {
		return nil
	}
	return list.Items
}

// discoverScalableResources finds CRD instances that support the scale subresource.
func discoverScalableResources(ctx context.Context, c client.Client, crds []apiextv1.CustomResourceDefinition, namespace string) []checks.ScalableResource {
	var scalable []checks.ScalableResource

	for i := range crds {
		crd := &crds[i]
		// Check if any version has the scale subresource
		hasScale := false
		var servedVersion string
		for _, v := range crd.Spec.Versions {
			if v.Subresources != nil && v.Subresources.Scale != nil && v.Served {
				hasScale = true
				if servedVersion == "" || v.Storage {
					servedVersion = v.Name
				}
			}
		}
		if !hasScale || servedVersion == "" {
			continue
		}

		gvk := schema.GroupVersionKind{
			Group:   crd.Spec.Group,
			Version: servedVersion,
			Kind:    crd.Spec.Names.ListKind,
		}
		gr := schema.GroupResource{
			Group:    crd.Spec.Group,
			Resource: crd.Spec.Names.Plural,
		}

		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(gvk)

		var opts []client.ListOption
		if crd.Spec.Scope == apiextv1.NamespaceScoped {
			opts = append(opts, client.InNamespace(namespace))
		}

		if err := c.List(ctx, list, opts...); err != nil {
			continue
		}

		for _, item := range list.Items {
			replicas, _, _ := unstructured.NestedInt64(item.Object, "spec", "replicas")
			scalable = append(scalable, checks.ScalableResource{
				Name:          item.GetName(),
				Namespace:     item.GetNamespace(),
				Replicas:      int32(replicas),
				GroupResource: gr,
			})
		}
	}

	return scalable
}

// helmChartMetadata is used to parse the chart metadata from Helm release secrets.
type helmChartMetadata struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type helmReleaseData struct {
	Chart struct {
		Metadata helmChartMetadata `json:"metadata"`
	} `json:"chart"`
}

// parseHelmRelease extracts chart name and version from a Helm release secret.
func parseHelmRelease(secret *corev1.Secret) (checks.HelmChartRelease, bool) {
	data, ok := secret.Data["release"]
	if !ok {
		return checks.HelmChartRelease{}, false
	}

	// Helm release data is base64-encoded then gzip-compressed JSON.
	// The secret data is already base64-decoded by Kubernetes.
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return checks.HelmChartRelease{}, false
	}
	defer func() { _ = reader.Close() }()

	decoded, err := io.ReadAll(reader)
	if err != nil {
		return checks.HelmChartRelease{}, false
	}

	var rel helmReleaseData
	if err := json.Unmarshal(decoded, &rel); err != nil {
		return checks.HelmChartRelease{}, false
	}

	if rel.Chart.Metadata.Name == "" {
		return checks.HelmChartRelease{}, false
	}

	return checks.HelmChartRelease{
		Name:      rel.Chart.Metadata.Name,
		Namespace: secret.Namespace,
		Version:   rel.Chart.Metadata.Version,
	}, true
}

// extractOpenshiftVersion gets the version string from the ClusterVersion status.
func extractOpenshiftVersion(cv *configv1.ClusterVersion) string {
	for _, h := range cv.Status.History {
		if h.State == configv1.CompletedUpdate {
			return h.Version
		}
	}
	if len(cv.Status.History) > 0 {
		return cv.Status.History[0].Version
	}
	return ""
}

// deriveOCPStatus determines the lifecycle status from the ClusterVersion.
func deriveOCPStatus(cv *configv1.ClusterVersion) string {
	version := extractOpenshiftVersion(cv)
	if version == "" {
		return ""
	}

	// Check conditions for upgrade/degraded status
	for _, cond := range cv.Status.Conditions {
		if cond.Type == "Progressing" && cond.Status == configv1.ConditionTrue {
			return "PreGA"
		}
	}

	// Parse version to determine lifecycle
	parts := strings.SplitN(version, ".", 3)
	if len(parts) < 2 {
		return "GA"
	}

	return "GA"
}
