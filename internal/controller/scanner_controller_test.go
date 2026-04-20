package controller

import (
	"context"
	"os"
	"testing"
	"time"

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
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	checksall "github.com/redhat-best-practices-for-k8s/checks/all"
	bpsv1alpha1 "github.com/redhat-best-practices-for-k8s/checks-types/api/v1alpha1"
)

func TestMain(m *testing.M) {
	checksall.Register()
	os.Exit(m.Run())
}

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

func newReconciler(c client.Client, s *runtime.Scheme) *ScannerReconciler {
	return &ScannerReconciler{Client: c, Scheme: s, Recorder: record.NewFakeRecorder(10)}
}

func TestReconcile_ScannerNotFound(t *testing.T) {
	s := newScheme()
	c := fake.NewClientBuilder().WithScheme(s).Build()

	r := newReconciler(c, s)
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

	r := newReconciler(c, s)
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
			Checks: []string{"access-control-pod-host-network"},
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

	r := newReconciler(c, s)
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
	if result.Spec.CheckName != "access-control-pod-host-network" {
		t.Errorf("expected check name access-control-pod-host-network, got %s", result.Spec.CheckName)
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
			Checks: []string{"access-control-pod-host-network"},
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

	r := newReconciler(c, s)
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
			Checks:       []string{"access-control-pod-host-network"},
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

	r := newReconciler(c, s)
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

func TestReconcile_InvalidScanInterval(t *testing.T) {
	s := newScheme()

	scanner := &bpsv1alpha1.BestPracticeScanner{
		ObjectMeta: metav1.ObjectMeta{Name: "scanner", Namespace: "ns"},
		Spec: bpsv1alpha1.BestPracticeScannerSpec{
			ScanInterval: "invalid-duration",
			Checks:       []string{"access-control-pod-host-network"},
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

	r := newReconciler(c, s)
	result, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "scanner", Namespace: "ns"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Invalid interval treated as one-shot, no requeue
	if result.RequeueAfter != 0 {
		t.Error("expected no requeue for invalid scanInterval (treated as one-shot)")
	}
}

func TestReconcile_EnforceUniqueness(t *testing.T) {
	s := newScheme()

	earlier := metav1.NewTime(time.Now().Add(-1 * time.Hour))
	later := metav1.NewTime(time.Now())

	existingScanner := &bpsv1alpha1.BestPracticeScanner{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "existing-scanner",
			Namespace:         "ns",
			CreationTimestamp: earlier,
		},
		Spec: bpsv1alpha1.BestPracticeScannerSpec{
			Checks: []string{"access-control-pod-host-network"},
		},
	}

	newScanner := &bpsv1alpha1.BestPracticeScanner{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "new-scanner",
			Namespace:         "ns",
			CreationTimestamp: later,
		},
		Spec: bpsv1alpha1.BestPracticeScannerSpec{
			Checks: []string{"access-control-pod-host-network"},
		},
	}

	c := fake.NewClientBuilder().WithScheme(s).
		WithObjects(existingScanner, newScanner).
		WithStatusSubresource(existingScanner, newScanner).
		Build()

	r := newReconciler(c, s)
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "new-scanner", Namespace: "ns"},
	})
	if err == nil {
		t.Fatal("expected error for duplicate scanner")
	}

	// Verify the newer scanner was set to Error phase
	var updated bpsv1alpha1.BestPracticeScanner
	_ = c.Get(context.Background(), types.NamespacedName{Name: "new-scanner", Namespace: "ns"}, &updated)
	if updated.Status.Phase != bpsv1alpha1.PhaseError {
		t.Errorf("expected Error phase, got %s", updated.Status.Phase)
	}
}

func TestReconcile_DeleteStaleResults(t *testing.T) {
	s := newScheme()

	scanner := &bpsv1alpha1.BestPracticeScanner{
		ObjectMeta: metav1.ObjectMeta{Name: "scanner", Namespace: "ns"},
		Spec: bpsv1alpha1.BestPracticeScannerSpec{
			Checks: []string{"access-control-pod-host-network"},
		},
	}

	// Pre-existing stale result from a check that is no longer in the scan
	staleResult := &bpsv1alpha1.BestPracticeResult{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "scanner.old-check",
			Namespace: "ns",
		},
		Spec: bpsv1alpha1.BestPracticeResultSpec{
			ScannerRef: "scanner",
			CheckName:  "old-check",
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "ns"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c1", Image: "img"}}},
	}

	c := fake.NewClientBuilder().WithScheme(s).
		WithObjects(scanner, staleResult, pod).
		WithStatusSubresource(scanner).
		WithIndex(&bpsv1alpha1.BestPracticeResult{}, "spec.scannerRef", func(obj client.Object) []string {
			result := obj.(*bpsv1alpha1.BestPracticeResult)
			return []string{result.Spec.ScannerRef}
		}).
		Build()

	r := newReconciler(c, s)
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "scanner", Namespace: "ns"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the stale result was deleted
	var resultList bpsv1alpha1.BestPracticeResultList
	_ = c.List(context.Background(), &resultList, client.InNamespace("ns"))

	for _, r := range resultList.Items {
		if r.Name == "scanner.old-check" {
			t.Error("stale result scanner.old-check should have been deleted")
		}
	}

	// Verify the current check result exists
	found := false
	for _, r := range resultList.Items {
		if r.Spec.CheckName == "access-control-pod-host-network" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected result for access-control-pod-host-network to exist")
	}
}

func TestReconcile_CompletedOneShotSkipsRescan(t *testing.T) {
	s := newScheme()

	now := metav1.Now()
	scanner := &bpsv1alpha1.BestPracticeScanner{
		ObjectMeta: metav1.ObjectMeta{Name: "scanner", Namespace: "ns"},
		Spec: bpsv1alpha1.BestPracticeScannerSpec{
			Checks: []string{"access-control-pod-host-network"},
		},
		Status: bpsv1alpha1.BestPracticeScannerStatus{
			Phase:        bpsv1alpha1.PhaseCompleted,
			LastScanTime: &now,
		},
	}

	c := fake.NewClientBuilder().WithScheme(s).
		WithObjects(scanner).
		WithStatusSubresource(scanner).
		Build()

	r := newReconciler(c, s)
	result, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "scanner", Namespace: "ns"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Error("expected no requeue for completed one-shot scan")
	}
}

func TestReconcile_CompletedPeriodicRequeuesWhenNotDue(t *testing.T) {
	s := newScheme()

	now := metav1.Now()
	scanner := &bpsv1alpha1.BestPracticeScanner{
		ObjectMeta: metav1.ObjectMeta{Name: "scanner", Namespace: "ns"},
		Spec: bpsv1alpha1.BestPracticeScannerSpec{
			ScanInterval: "1h",
			Checks:       []string{"access-control-pod-host-network"},
		},
		Status: bpsv1alpha1.BestPracticeScannerStatus{
			Phase:        bpsv1alpha1.PhaseCompleted,
			LastScanTime: &now,
		},
	}

	c := fake.NewClientBuilder().WithScheme(s).
		WithObjects(scanner).
		WithStatusSubresource(scanner).
		Build()

	r := newReconciler(c, s)
	result, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "scanner", Namespace: "ns"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Error("expected requeue for periodic scan not yet due")
	}
	if result.RequeueAfter > 1*time.Hour {
		t.Errorf("expected requeue within 1h, got %v", result.RequeueAfter)
	}
}

func TestReconcile_CatalogURL(t *testing.T) {
	s := newScheme()

	scanner := &bpsv1alpha1.BestPracticeScanner{
		ObjectMeta: metav1.ObjectMeta{Name: "scanner", Namespace: "ns"},
		Spec: bpsv1alpha1.BestPracticeScannerSpec{
			Checks: []string{"access-control-pod-host-network"},
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

	r := &ScannerReconciler{
		Client:         c,
		Scheme:         s,
		Recorder:       record.NewFakeRecorder(10),
		CatalogURLBase: "https://example.com/CATALOG.md",
	}
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

	result := resultList.Items[0]
	if result.Spec.CatalogURL == "" {
		t.Error("expected CatalogURL to be set when CatalogURLBase is configured")
	}
}

func TestReconcile_DefaultTargetNamespace(t *testing.T) {
	s := newScheme()

	scanner := &bpsv1alpha1.BestPracticeScanner{
		ObjectMeta: metav1.ObjectMeta{Name: "scanner", Namespace: "my-ns"},
		Spec: bpsv1alpha1.BestPracticeScannerSpec{
			Checks: []string{"access-control-pod-host-network"},
		},
	}

	// Pod in scanner's namespace (should be discovered when targetNamespace is empty)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "my-ns"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c1", Image: "img"}}},
	}

	c := fake.NewClientBuilder().WithScheme(s).
		WithObjects(scanner, pod).
		WithStatusSubresource(scanner).
		Build()

	r := newReconciler(c, s)
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "scanner", Namespace: "my-ns"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify results were created in scanner's namespace
	var resultList bpsv1alpha1.BestPracticeResultList
	_ = c.List(context.Background(), &resultList, client.InNamespace("my-ns"))
	if len(resultList.Items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resultList.Items))
	}
}

func TestReconcile_SuspendAlreadyIdle(t *testing.T) {
	s := newScheme()

	scanner := &bpsv1alpha1.BestPracticeScanner{
		ObjectMeta: metav1.ObjectMeta{Name: "scanner", Namespace: "ns"},
		Spec:       bpsv1alpha1.BestPracticeScannerSpec{Suspend: true},
		Status: bpsv1alpha1.BestPracticeScannerStatus{
			Phase: bpsv1alpha1.PhaseIdle,
		},
	}

	c := fake.NewClientBuilder().WithScheme(s).
		WithObjects(scanner).
		WithStatusSubresource(scanner).
		Build()

	r := newReconciler(c, s)
	result, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "scanner", Namespace: "ns"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Error("expected no requeue for suspended scanner")
	}
}
