package probe

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestEnsureDaemonSet_Create(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	err := EnsureDaemonSet(context.Background(), client, "test-ns")
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
