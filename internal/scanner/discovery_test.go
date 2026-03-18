package scanner

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestDiscover(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = policyv1.AddToScheme(scheme)
	_ = storagev1.AddToScheme(scheme)
	_ = apiextv1.AddToScheme(scheme)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "test-ns"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c1", Image: "img"}}},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "test-svc", Namespace: "test-ns"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 80}}},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod, svc).Build()

	resources, err := Discover(context.Background(), client, "test-ns", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources.Pods) != 1 {
		t.Errorf("expected 1 pod, got %d", len(resources.Pods))
	}
	if len(resources.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(resources.Services))
	}
	if resources.Namespaces[0] != "test-ns" {
		t.Errorf("expected namespace test-ns, got %s", resources.Namespaces[0])
	}
}

func TestDiscover_WithLabelSelector(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = policyv1.AddToScheme(scheme)
	_ = storagev1.AddToScheme(scheme)
	_ = apiextv1.AddToScheme(scheme)

	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c1", Image: "img"}}},
	}
	pod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "ns", Labels: map[string]string{"app": "db"}},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c1", Image: "img"}}},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod1, pod2).Build()

	ls := &metav1.LabelSelector{MatchLabels: map[string]string{"app": "web"}}
	resources, err := Discover(context.Background(), client, "ns", ls, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources.Pods) != 1 {
		t.Errorf("expected 1 pod, got %d", len(resources.Pods))
	}
	if resources.Pods[0].Name != "pod1" {
		t.Errorf("expected pod1, got %s", resources.Pods[0].Name)
	}
}

func TestDiscover_Roles(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = policyv1.AddToScheme(scheme)
	_ = storagev1.AddToScheme(scheme)
	_ = apiextv1.AddToScheme(scheme)

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: "test-role", Namespace: "test-ns"},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(role).Build()

	resources, err := Discover(context.Background(), client, "test-ns", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources.Roles) != 1 {
		t.Errorf("expected 1 role, got %d", len(resources.Roles))
	}
	if resources.Roles[0].Name != "test-role" {
		t.Errorf("expected test-role, got %s", resources.Roles[0].Name)
	}
}

func TestDiscover_HelmChartReleases(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = policyv1.AddToScheme(scheme)
	_ = storagev1.AddToScheme(scheme)
	_ = apiextv1.AddToScheme(scheme)

	// Create a helm release secret with gzipped JSON data
	releaseData := helmReleaseData{}
	releaseData.Chart.Metadata.Name = "my-chart"
	releaseData.Chart.Metadata.Version = "1.2.3"
	jsonData, _ := json.Marshal(releaseData)

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write(jsonData)
	_ = gz.Close()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "sh.helm.release.v1.my-chart.v1", Namespace: "test-ns"},
		Type:       "helm.sh/release.v1",
		Data: map[string][]byte{
			"release": buf.Bytes(),
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret).Build()

	resources, err := Discover(context.Background(), client, "test-ns", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources.HelmChartReleases) != 1 {
		t.Fatalf("expected 1 helm chart release, got %d", len(resources.HelmChartReleases))
	}
	if resources.HelmChartReleases[0].Name != "my-chart" {
		t.Errorf("expected chart name my-chart, got %s", resources.HelmChartReleases[0].Name)
	}
	if resources.HelmChartReleases[0].Version != "1.2.3" {
		t.Errorf("expected chart version 1.2.3, got %s", resources.HelmChartReleases[0].Version)
	}
}

func TestDiscover_GracefulSkipUnregisteredCRDs(t *testing.T) {
	// Use a scheme that does NOT have OpenShift/OLM types registered.
	// Discovery should succeed gracefully, just with empty results for those types.
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = policyv1.AddToScheme(scheme)
	_ = storagev1.AddToScheme(scheme)
	_ = apiextv1.AddToScheme(scheme)

	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	resources, err := Discover(context.Background(), client, "test-ns", nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resources.ClusterVersion != nil {
		t.Error("expected nil ClusterVersion on vanilla K8s")
	}
	if len(resources.CSVs) != 0 {
		t.Errorf("expected 0 CSVs, got %d", len(resources.CSVs))
	}
	if len(resources.NetworkAttachmentDefinitions) != 0 {
		t.Errorf("expected 0 NADs, got %d", len(resources.NetworkAttachmentDefinitions))
	}
}
