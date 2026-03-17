package probe

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestEnsureDaemonSet_Create(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	err := EnsureDaemonSet(context.Background(), client, "test-ns", ProbeImage)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var ds appsv1.DaemonSet
	err = client.Get(context.Background(), types.NamespacedName{Name: ProbeName, Namespace: "test-ns"}, &ds)
	if err != nil {
		t.Fatalf("DaemonSet not found: %v", err)
	}

	if ds.Spec.Template.Spec.HostNetwork != true {
		t.Error("expected HostNetwork to be true")
	}
	if ds.Spec.Template.Spec.Containers[0].Image != ProbeImage {
		t.Errorf("expected image %s, got %s", ProbeImage, ds.Spec.Template.Spec.Containers[0].Image)
	}
}

func TestEnsureDaemonSet_CustomImage(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	customImage := "my-registry.io/custom-probe:v1.0.0"
	err := EnsureDaemonSet(context.Background(), client, "test-ns", customImage)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var ds appsv1.DaemonSet
	err = client.Get(context.Background(), types.NamespacedName{Name: ProbeName, Namespace: "test-ns"}, &ds)
	if err != nil {
		t.Fatalf("DaemonSet not found: %v", err)
	}

	if ds.Spec.Template.Spec.Containers[0].Image != customImage {
		t.Errorf("expected custom image %s, got %s", customImage, ds.Spec.Template.Spec.Containers[0].Image)
	}
}

func TestEnsureDaemonSet_EmptyImageUsesDefault(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	err := EnsureDaemonSet(context.Background(), client, "test-ns", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var ds appsv1.DaemonSet
	err = client.Get(context.Background(), types.NamespacedName{Name: ProbeName, Namespace: "test-ns"}, &ds)
	if err != nil {
		t.Fatalf("DaemonSet not found: %v", err)
	}

	if ds.Spec.Template.Spec.Containers[0].Image != ProbeImage {
		t.Errorf("expected default image %s when empty string provided, got %s", ProbeImage, ds.Spec.Template.Spec.Containers[0].Image)
	}
}

func TestEnsureDaemonSet_NoUpdateWhenUnchanged(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	// Create initial DaemonSet
	err := EnsureDaemonSet(context.Background(), client, "test-ns", ProbeImage)
	if err != nil {
		t.Fatalf("unexpected error on create: %v", err)
	}

	// Get the DaemonSet and record its resource version
	var ds appsv1.DaemonSet
	err = client.Get(context.Background(), types.NamespacedName{Name: ProbeName, Namespace: "test-ns"}, &ds)
	if err != nil {
		t.Fatalf("DaemonSet not found: %v", err)
	}
	initialResourceVersion := ds.ResourceVersion

	// Call EnsureDaemonSet again with unchanged spec
	err = EnsureDaemonSet(context.Background(), client, "test-ns", ProbeImage)
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}

	// Get the DaemonSet again and verify resource version hasn't changed
	err = client.Get(context.Background(), types.NamespacedName{Name: ProbeName, Namespace: "test-ns"}, &ds)
	if err != nil {
		t.Fatalf("DaemonSet not found after second call: %v", err)
	}

	if ds.ResourceVersion != initialResourceVersion {
		t.Errorf("ResourceVersion changed from %s to %s, indicating unnecessary update",
			initialResourceVersion, ds.ResourceVersion)
	}
}

func TestMapProbePods(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "probe-node1",
			Namespace: "test-ns",
			Labels:    map[string]string{ProbeLabel: ProbeLabelVal},
		},
		Spec:   corev1.PodSpec{NodeName: "node1"},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).Build()

	probeMap, err := MapProbePods(context.Background(), client, "test-ns")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(probeMap) != 1 {
		t.Fatalf("expected 1 probe pod, got %d", len(probeMap))
	}
	if probeMap["node1"] == nil {
		t.Error("expected probe pod for node1")
	}
}

func TestNewExecutor_DefaultTimeout(t *testing.T) {
	config := &rest.Config{}
	executor, err := NewExecutor(config, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if executor.timeout != DefaultExecTimeout {
		t.Errorf("expected default timeout %v, got %v", DefaultExecTimeout, executor.timeout)
	}
}

func TestNewExecutor_CustomTimeout(t *testing.T) {
	config := &rest.Config{}
	customTimeout := 60 * time.Second
	executor, err := NewExecutor(config, customTimeout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if executor.timeout != customTimeout {
		t.Errorf("expected custom timeout %v, got %v", customTimeout, executor.timeout)
	}
}
