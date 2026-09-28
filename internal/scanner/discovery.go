package scanner

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"

	netattdefv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	apiserverv1 "github.com/openshift/api/apiserver/v1"
	configv1 "github.com/openshift/api/config/v1"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmpackagev1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
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
	"golang.org/x/sync/errgroup"
)

const maxConcurrentDiscoveryCalls = 8

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

	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(maxConcurrentDiscoveryCalls)
	listRequired := func(list client.ObjectList, opts ...client.ListOption) {
		group.Go(func() error {
			return c.List(groupCtx, list, opts...)
		})
	}
	listOptionalInGroup := func(list client.ObjectList, onSuccess func(), opts ...client.ListOption) {
		group.Go(func() error {
			if listOptional(groupCtx, c, list, opts...) {
				onSuccess()
			}
			return nil
		})
	}

	// --- Core K8s resources (required) ---
	var pods corev1.PodList
	listRequired(&pods, labelOpts...)
	var services corev1.ServiceList
	listRequired(&services, nsOpts...)
	var serviceAccounts corev1.ServiceAccountList
	listRequired(&serviceAccounts, nsOpts...)
	var roles rbacv1.RoleList
	listRequired(&roles, nsOpts...)
	var roleBindings rbacv1.RoleBindingList
	listRequired(&roleBindings, nsOpts...)
	var crbs rbacv1.ClusterRoleBindingList
	listRequired(&crbs)
	var crds apiextv1.CustomResourceDefinitionList
	listRequired(&crds)
	var deployments appsv1.DeploymentList
	listRequired(&deployments, nsOpts...)
	var statefulSets appsv1.StatefulSetList
	listRequired(&statefulSets, nsOpts...)
	var daemonSets appsv1.DaemonSetList
	listRequired(&daemonSets, nsOpts...)
	var netPolicies networkingv1.NetworkPolicyList
	listRequired(&netPolicies, nsOpts...)
	var quotas corev1.ResourceQuotaList
	listRequired(&quotas, nsOpts...)
	var pdbs policyv1.PodDisruptionBudgetList
	listRequired(&pdbs, nsOpts...)
	var nodes corev1.NodeList
	listRequired(&nodes)
	var pvs corev1.PersistentVolumeList
	listRequired(&pvs)
	var scs storagev1.StorageClassList
	listRequired(&scs)
	var secrets corev1.SecretList
	listRequired(&secrets, nsOpts...)

	// K8s version
	if discoveryClient != nil {
		group.Go(func() error {
			if info, err := discoveryClient.ServerVersion(); err == nil {
				resources.K8sVersion = info.GitVersion
			}
			return nil
		})
	}

	// --- OpenShift-specific resources (optional — graceful skip) ---
	var cv configv1.ClusterVersion
	group.Go(func() error {
		if err := c.Get(groupCtx, types.NamespacedName{Name: "version"}, &cv); err == nil {
			resources.ClusterVersion = &cv
			resources.OpenshiftVersion = extractOpenshiftVersion(&cv)
			resources.OCPStatus = deriveOCPStatus(&cv, resources.OpenshiftVersion)
		}
		return nil
	})
	var cos configv1.ClusterOperatorList
	listOptionalInGroup(&cos, func() { resources.ClusterOperators = cos.Items })
	var arcs apiserverv1.APIRequestCountList
	listOptionalInGroup(&arcs, func() { resources.APIRequestCounts = arcs.Items })

	// --- OLM resources (optional — graceful skip) ---
	var csvs olmv1alpha1.ClusterServiceVersionList
	listOptionalInGroup(&csvs, func() { resources.CSVs = csvs.Items }, nsOpts...)
	var catalogs olmv1alpha1.CatalogSourceList
	listOptionalInGroup(&catalogs, func() { resources.CatalogSources = catalogs.Items })
	var subs olmv1alpha1.SubscriptionList
	listOptionalInGroup(&subs, func() { resources.Subscriptions = subs.Items }, nsOpts...)
	var pkgs olmpackagev1.PackageManifestList
	listOptionalInGroup(&pkgs, func() { resources.PackageManifests = pkgs.Items }, nsOpts...)

	// --- Networking resources (optional — graceful skip) ---
	var nads netattdefv1.NetworkAttachmentDefinitionList
	listOptionalInGroup(&nads, func() { resources.NetworkAttachmentDefinitions = nads.Items }, nsOpts...)

	// SR-IOV resources (unstructured)
	group.Go(func() error {
		resources.SriovNetworks = listUnstructured(groupCtx, c, schema.GroupVersionResource{
			Group: "sriovnetwork.openshift.io", Version: "v1", Resource: "sriovnetworks",
		}, namespace)
		return nil
	})
	group.Go(func() error {
		resources.SriovNetworkNodePolicies = listUnstructured(groupCtx, c, schema.GroupVersionResource{
			Group: "sriovnetwork.openshift.io", Version: "v1", Resource: "sriovnetworknodepolicies",
		}, "")
		return nil
	})

	if err := group.Wait(); err != nil {
		return nil, err
	}

	resources.Pods = pods.Items
	resources.Services = services.Items
	resources.ServiceAccounts = serviceAccounts.Items
	resources.Roles = roles.Items
	resources.RoleBindings = roleBindings.Items
	resources.ClusterRoleBindings = crbs.Items
	resources.CRDs = crds.Items
	resources.Deployments = deployments.Items
	resources.StatefulSets = statefulSets.Items
	resources.DaemonSets = daemonSets.Items
	resources.NetworkPolicies = netPolicies.Items
	resources.ResourceQuotas = quotas.Items
	resources.PodDisruptionBudgets = pdbs.Items
	resources.Nodes = nodes.Items
	resources.PersistentVolumes = pvs.Items
	resources.StorageClasses = scs.Items

	// Helm chart releases (secrets with type helm.sh/release.v1)
	for i := range secrets.Items {
		if secrets.Items[i].Type == "helm.sh/release.v1" {
			if release, ok := parseHelmRelease(&secrets.Items[i]); ok {
				resources.HelmChartReleases = append(resources.HelmChartReleases, release)
			}
		}
	}

	// Scalable CRD resources depend on the CRDs list.
	resources.ScalableResources = discoverScalableResources(ctx, c, crds.Items, namespace)

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

	const maxHelmReleaseSize = 10 << 20 // 10 MiB
	var rel helmReleaseData
	if err := json.NewDecoder(io.LimitReader(reader, maxHelmReleaseSize)).Decode(&rel); err != nil {
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
