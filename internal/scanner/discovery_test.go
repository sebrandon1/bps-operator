package scanner

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestDiscover(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
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

	resources, err := Discover(context.Background(), client, "test-ns", nil)
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
	resources, err := Discover(context.Background(), client, "ns", ls)
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
