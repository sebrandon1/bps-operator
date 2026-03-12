package controller

import (
	"context"
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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	bpsv1alpha1 "github.com/sebrandon1/bps-operator/api/v1alpha1"

	// Register checks
	_ "github.com/redhat-best-practices-for-k8s/checks/accesscontrol"
	_ "github.com/redhat-best-practices-for-k8s/checks/observability"
)

func newScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = rbacv1.AddToScheme(s)
	_ = appsv1.AddToScheme(s)
	_ = networkingv1.AddToScheme(s)
	_ = policyv1.AddToScheme(s)
	_ = storagev1.AddToScheme(s)
	_ = apiextv1.AddToScheme(s)
	_ = bpsv1alpha1.AddToScheme(s)
	return s
}

func TestReconcile_ScannerNotFound(t *testing.T) {
	s := newScheme()
	c := fake.NewClientBuilder().WithScheme(s).Build()

	r := &ScannerReconciler{Client: c, Scheme: s}
	result, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "missing", Namespace: "test"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Error("unexpected requeue")
	}
}

func TestReconcile_Suspend(t *testing.T) {
	s := newScheme()
	scanner := &bpsv1alpha1.BestPracticeScanner{
		ObjectMeta: metav1.ObjectMeta{Name: "test-scanner", Namespace: "test-ns"},
		Spec:       bpsv1alpha1.BestPracticeScannerSpec{Suspend: true},
	}

	c := fake.NewClientBuilder().WithScheme(s).
		WithObjects(scanner).
		WithStatusSubresource(scanner).
		Build()

	r := &ScannerReconciler{Client: c, Scheme: s}
	result, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "test-scanner", Namespace: "test-ns"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Error("unexpected requeue")
	}

	// Verify phase is Idle
	var updated bpsv1alpha1.BestPracticeScanner
	_ = c.Get(context.Background(), types.NamespacedName{Name: "test-scanner", Namespace: "test-ns"}, &updated)
	if updated.Status.Phase != bpsv1alpha1.PhaseIdle {
		t.Errorf("expected Idle, got %s", updated.Status.Phase)
	}
}

func TestReconcile_FullScan(t *testing.T) {
	s := newScheme()

	scanner := &bpsv1alpha1.BestPracticeScanner{
		ObjectMeta: metav1.ObjectMeta{Name: "scanner", Namespace: "ns"},
		Spec: bpsv1alpha1.BestPracticeScannerSpec{
			Checks: []string{"access-control-host-network"},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "good-pod", Namespace: "ns"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "c1", Image: "img"}},
		},
	}

	c := fake.NewClientBuilder().WithScheme(s).
		WithObjects(scanner, pod).
		WithStatusSubresource(scanner).
		Build()

	r := &ScannerReconciler{Client: c, Scheme: s}
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "scanner", Namespace: "ns"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify result was created
	var resultList bpsv1alpha1.BestPracticeResultList
	_ = c.List(context.Background(), &resultList, client.InNamespace("ns"))
	if len(resultList.Items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resultList.Items))
	}

	result := resultList.Items[0]
	if result.Spec.CheckName != "access-control-host-network" {
		t.Errorf("expected check name access-control-host-network, got %s", result.Spec.CheckName)
	}
	if result.Spec.ComplianceStatus != bpsv1alpha1.StatusCompliant {
		t.Errorf("expected Compliant, got %s", result.Spec.ComplianceStatus)
	}

	// Verify scanner status
	var updated bpsv1alpha1.BestPracticeScanner
	_ = c.Get(context.Background(), types.NamespacedName{Name: "scanner", Namespace: "ns"}, &updated)
	if updated.Status.Phase != bpsv1alpha1.PhaseCompleted {
		t.Errorf("expected Completed, got %s", updated.Status.Phase)
	}
	if updated.Status.Summary == nil {
		t.Fatal("expected summary to be set")
	}
	if updated.Status.Summary.Total != 1 {
		t.Errorf("expected total 1, got %d", updated.Status.Summary.Total)
	}
	if updated.Status.Summary.Compliant != 1 {
		t.Errorf("expected compliant 1, got %d", updated.Status.Summary.Compliant)
	}
}

func TestReconcile_NonCompliant(t *testing.T) {
	s := newScheme()

	scanner := &bpsv1alpha1.BestPracticeScanner{
		ObjectMeta: metav1.ObjectMeta{Name: "scanner", Namespace: "ns"},
		Spec: bpsv1alpha1.BestPracticeScannerSpec{
			Checks: []string{"access-control-host-network"},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "bad-pod", Namespace: "ns"},
		Spec: corev1.PodSpec{
			HostNetwork: true,
			Containers:  []corev1.Container{{Name: "c1", Image: "img"}},
		},
	}

	c := fake.NewClientBuilder().WithScheme(s).
		WithObjects(scanner, pod).
		WithStatusSubresource(scanner).
		Build()

	r := &ScannerReconciler{Client: c, Scheme: s}
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "scanner", Namespace: "ns"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resultList bpsv1alpha1.BestPracticeResultList
	_ = c.List(context.Background(), &resultList, client.InNamespace("ns"))
	if len(resultList.Items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resultList.Items))
	}
	if resultList.Items[0].Spec.ComplianceStatus != bpsv1alpha1.StatusNonCompliant {
		t.Errorf("expected NonCompliant, got %s", resultList.Items[0].Spec.ComplianceStatus)
	}
}

func TestReconcile_WithScanInterval(t *testing.T) {
	s := newScheme()

	scanner := &bpsv1alpha1.BestPracticeScanner{
		ObjectMeta: metav1.ObjectMeta{Name: "scanner", Namespace: "ns"},
		Spec: bpsv1alpha1.BestPracticeScannerSpec{
			ScanInterval: "10m",
			Checks:       []string{"access-control-host-network"},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "ns"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c1", Image: "img"}}},
	}

	c := fake.NewClientBuilder().WithScheme(s).
		WithObjects(scanner, pod).
		WithStatusSubresource(scanner).
		Build()

	r := &ScannerReconciler{Client: c, Scheme: s}
	result, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "scanner", Namespace: "ns"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Error("expected RequeueAfter to be set")
	}
}
