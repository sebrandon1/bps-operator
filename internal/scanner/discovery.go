package scanner

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/redhat-best-practices-for-k8s/checks"
)

// listOptional lists resources and silently returns false if the API is not registered (for OpenShift/OLM types).
func listOptional(ctx context.Context, c client.Client, list client.ObjectList, opts ...client.ListOption) bool {
	err := c.List(ctx, list, opts...)
	return err == nil
}

// Discover lists all relevant resources in the target namespace.
func Discover(ctx context.Context, c client.Client, namespace string, labelSelector *metav1.LabelSelector, discoveryClient discovery.ServerVersionInterface) (*checks.DiscoveredResources, error) {
	resources := &checks.DiscoveredResources{
		Namespaces: []string{namespace},
	}

	// Build list options
	labelOpts := []client.ListOption{client.InNamespace(namespace)}
	if labelSelector != nil {
		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return nil, err
		}
		labelOpts = append(labelOpts, client.MatchingLabelsSelector{Selector: selector})
	}
	nsOpts := []client.ListOption{client.InNamespace(namespace)}

	// --- Core K8s resources (required) ---

	var pods corev1.PodList
	if err := c.List(ctx, &pods, labelOpts...); err != nil {
		return nil, err
	}
	resources.Pods = pods.Items

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

	var crbs rbacv1.ClusterRoleBindingList
	if err := c.List(ctx, &crbs); err != nil {
		return nil, err
	}
	resources.ClusterRoleBindings = crbs.Items

	var crds apiextv1.CustomResourceDefinitionList
	if err := c.List(ctx, &crds); err != nil {
		return nil, err
	}
	resources.CRDs = crds.Items

	var deployments appsv1.DeploymentList
	if err := c.List(ctx, &deployments, nsOpts...); err != nil {
		return nil, err
	}
	resources.Deployments = deployments.Items

	var statefulSets appsv1.StatefulSetList
	if err := c.List(ctx, &statefulSets, nsOpts...); err != nil {
		return nil, err
	}
	resources.StatefulSets = statefulSets.Items

	var daemonSets appsv1.DaemonSetList
	if err := c.List(ctx, &daemonSets, nsOpts...); err != nil {
		return nil, err
	}
	resources.DaemonSets = daemonSets.Items

	var netPolicies networkingv1.NetworkPolicyList
	if err := c.List(ctx, &netPolicies, nsOpts...); err != nil {
		return nil, err
	}
	resources.NetworkPolicies = netPolicies.Items

	var quotas corev1.ResourceQuotaList
	if err := c.List(ctx, &quotas, nsOpts...); err != nil {
		return nil, err
	}
	resources.ResourceQuotas = quotas.Items

	var pdbs policyv1.PodDisruptionBudgetList
	if err := c.List(ctx, &pdbs, nsOpts...); err != nil {
		return nil, err
	}
	resources.PodDisruptionBudgets = pdbs.Items

	var nodes corev1.NodeList
	if err := c.List(ctx, &nodes); err != nil {
		return nil, err
	}
	resources.Nodes = nodes.Items

	var pvs corev1.PersistentVolumeList
	if err := c.List(ctx, &pvs); err != nil {
		return nil, err
	}
	resources.PersistentVolumes = pvs.Items

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

	// --- OpenShift-specific resources (optional — graceful skip) ---

	var cv configv1.ClusterVersion
	if err := c.Get(ctx, types.NamespacedName{Name: "version"}, &cv); err == nil {
		resources.ClusterVersion = &cv
		resources.OpenshiftVersion = extractOpenshiftVersion(&cv)
		resources.OCPStatus = deriveOCPStatus(&cv, resources.OpenshiftVersion)
	}

	var cos configv1.ClusterOperatorList
	if listOptional(ctx, c, &cos) {
		resources.ClusterOperators = cos.Items
	}

	var arcs apiserverv1.APIRequestCountList
	if listOptional(ctx, c, &arcs) {
		resources.APIRequestCounts = arcs.Items
	}

	// --- OLM resources (optional — graceful skip) ---

	var csvs olmv1alpha1.ClusterServiceVersionList
	if listOptional(ctx, c, &csvs, nsOpts...) {
		resources.CSVs = csvs.Items
	}

	var catalogs olmv1alpha1.CatalogSourceList
	if listOptional(ctx, c, &catalogs) {
		resources.CatalogSources = catalogs.Items
	}

	var subs olmv1alpha1.SubscriptionList
	if listOptional(ctx, c, &subs, nsOpts...) {
		resources.Subscriptions = subs.Items
	}

	var pkgs olmpackagev1.PackageManifestList
	if listOptional(ctx, c, &pkgs, nsOpts...) {
		resources.PackageManifests = pkgs.Items
	}

	// --- Networking resources (optional — graceful skip) ---

	var nads netattdefv1.NetworkAttachmentDefinitionList
	if listOptional(ctx, c, &nads, nsOpts...) {
		resources.NetworkAttachmentDefinitions = nads.Items
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

// listUnstructured lists unstructured resources, returning empty on any error (graceful skip).
func listUnstructured(ctx context.Context, c client.Client, gvr schema.GroupVersionResource, namespace string) []unstructured.Unstructured {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gvr.Group,
		Version: gvr.Version,
		Kind:    gvr.Resource,
	})

	var opts []client.ListOption
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

type helmReleaseData struct {
	Chart struct {
		Metadata struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"metadata"`
	} `json:"chart"`
}

// parseHelmRelease extracts chart name and version from a Helm release secret.
func parseHelmRelease(secret *corev1.Secret) (checks.HelmChartRelease, bool) {
	data, ok := secret.Data["release"]
	if !ok {
		return checks.HelmChartRelease{}, false
	}

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
func deriveOCPStatus(cv *configv1.ClusterVersion, version string) string {
	if version == "" {
		return ""
	}

	for _, cond := range cv.Status.Conditions {
		if cond.Type == "Progressing" && cond.Status == configv1.ConditionTrue {
			return "PreGA"
		}
	}

	return "GA"
}
