package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/redhat-best-practices-for-k8s/checks"
	bpsv1alpha1 "github.com/sebrandon1/bps-operator/api/v1alpha1"
	"github.com/sebrandon1/bps-operator/internal/probe"
	"github.com/sebrandon1/bps-operator/internal/scanner"
)

// ScannerReconciler reconciles a BestPracticeScanner object.
type ScannerReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ProbeExecutor     checks.ProbeExecutor
	OperatorNamespace string
	ProbeImage        string
}

// +kubebuilder:rbac:groups=bps.openshift.io,resources=bestpracticescanners,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bps.openshift.io,resources=bestpracticescanners/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=bps.openshift.io,resources=bestpracticescanners/finalizers,verbs=update
// +kubebuilder:rbac:groups=bps.openshift.io,resources=bestpracticeresults,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods;services;serviceaccounts;namespaces;nodes;persistentvolumes;resourcequotas,verbs=get;list;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings;clusterrolebindings,verbs=get;list;watch
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=daemonsets;deployments;statefulsets,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=pods/exec,verbs=create
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch
// +kubebuilder:rbac:groups=storage.k8s.io,resources=storageclasses,verbs=get;list;watch

func (r *ScannerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the BestPracticeScanner CR
	var scannerCR bpsv1alpha1.BestPracticeScanner
	if err := r.Get(ctx, req.NamespacedName, &scannerCR); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Parse scanInterval once if specified
	var scanInterval time.Duration
	if scannerCR.Spec.ScanInterval != "" {
		var err error
		scanInterval, err = time.ParseDuration(scannerCR.Spec.ScanInterval)
		if err != nil {
			logger.Error(err, "Invalid scanInterval, treating as one-shot scan", "interval", scannerCR.Spec.ScanInterval)
			scanInterval = 0 // Treat as one-shot
		}
	}

	// Skip if scan already completed and not yet time for the next one.
	// This prevents a tight reconcile loop caused by the Owns watch on
	// BestPracticeResult objects (every create/update triggers re-reconcile).
	if scannerCR.Status.Phase == bpsv1alpha1.PhaseCompleted && scannerCR.Status.LastScanTime != nil {
		if scanInterval == 0 {
			// One-shot scan already done
			return ctrl.Result{}, nil
		}
		if time.Since(scannerCR.Status.LastScanTime.Time) < scanInterval {
			// Not yet time for the next periodic scan
			requeueIn := time.Until(scannerCR.Status.LastScanTime.Add(scanInterval))
			return ctrl.Result{RequeueAfter: requeueIn}, nil
		}
	}

	// Enforce one scanner per namespace
	var scannerList bpsv1alpha1.BestPracticeScannerList
	if err := r.List(ctx, &scannerList, client.InNamespace(scannerCR.Namespace)); err != nil {
		return ctrl.Result{}, err
	}
	for i := range scannerList.Items {
		other := &scannerList.Items[i]
		if other.Name != scannerCR.Name && other.CreationTimestamp.Before(&scannerCR.CreationTimestamp) {
			logger.Info("Another scanner already exists in this namespace, setting phase to Error",
				"existing", other.Name)
			scannerCR.Status.Phase = bpsv1alpha1.PhaseError
			if err := r.Status().Update(ctx, &scannerCR); err != nil {
				logger.Error(err, "Failed to update scanner status to Error")
			}
			return ctrl.Result{}, fmt.Errorf("namespace %s already has scanner %q; only one scanner per namespace is allowed", scannerCR.Namespace, other.Name)
		}
	}

	// Handle suspend
	if scannerCR.Spec.Suspend {
		if scannerCR.Status.Phase != bpsv1alpha1.PhaseIdle {
			scannerCR.Status.Phase = bpsv1alpha1.PhaseIdle
			if err := r.Status().Update(ctx, &scannerCR); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Determine target namespace
	targetNS := scannerCR.Spec.TargetNamespace
	if targetNS == "" {
		targetNS = scannerCR.Namespace
	}

	// Set phase to Scanning
	scannerCR.Status.Phase = bpsv1alpha1.PhaseScanning
	if err := r.Status().Update(ctx, &scannerCR); err != nil {
		return ctrl.Result{}, err
	}

	// Ensure probe DaemonSet
	if r.OperatorNamespace != "" {
		if err := probe.EnsureDaemonSet(ctx, r.Client, r.OperatorNamespace, r.ProbeImage); err != nil {
			logger.Error(err, "Failed to ensure probe DaemonSet")
		}
	}

	// Discover resources
	resources, err := scanner.Discover(ctx, r.Client, targetNS, scannerCR.Spec.LabelSelector)
	if err != nil {
		scannerCR.Status.Phase = bpsv1alpha1.PhaseError
		if updateErr := r.Status().Update(ctx, &scannerCR); updateErr != nil {
			logger.Error(updateErr, "Failed to update scanner status to Error")
		}
		return ctrl.Result{}, fmt.Errorf("discovering resources: %w", err)
	}

	// Map probe pods if available
	if r.OperatorNamespace != "" {
		probePods, err := probe.MapProbePods(ctx, r.Client, r.OperatorNamespace)
		if err != nil {
			logger.Error(err, "Failed to map probe pods, probe-based checks will be skipped")
		} else {
			resources.ProbePods = probePods
			resources.ProbeExecutor = r.ProbeExecutor
		}
	}

	// Run checks
	checksToRun := checks.Filtered(scannerCR.Spec.Checks)
	now := metav1.Now()
	summary := bpsv1alpha1.ScanSummary{Total: len(checksToRun)}
	resultNames := make(map[string]bool)

	for _, check := range checksToRun {
		checkResult := check.Fn(resources)

		// Count summary
		status := bpsv1alpha1.ComplianceStatus(checkResult.ComplianceStatus)
		switch status {
		case bpsv1alpha1.StatusCompliant:
			summary.Compliant++
		case bpsv1alpha1.StatusNonCompliant:
			summary.NonCompliant++
		case bpsv1alpha1.StatusError:
			summary.Error++
		case bpsv1alpha1.StatusSkipped:
			summary.Skipped++
		}

		// Upsert BestPracticeResult
		resultName := fmt.Sprintf("%s-%s", scannerCR.Name, check.Name)
		resultNames[resultName] = true

		result := &bpsv1alpha1.BestPracticeResult{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resultName,
				Namespace: scannerCR.Namespace,
			},
		}

		op, err := controllerutil.CreateOrUpdate(ctx, r.Client, result, func() error {
			// Set owner reference
			if err := controllerutil.SetControllerReference(&scannerCR, result, r.Scheme); err != nil {
				return err
			}

			// Convert check details to API details
			var apiDetails []bpsv1alpha1.ResourceDetail
			for _, d := range checkResult.Details {
				apiDetails = append(apiDetails, bpsv1alpha1.ResourceDetail{
					Kind:      d.Kind,
					Name:      d.Name,
					Namespace: d.Namespace,
					Compliant: d.Compliant,
					Message:   d.Message,
				})
			}

			var catalogURL string
			if check.CatalogID != "" {
				catalogURL = "https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#" + check.CatalogID
			}

			result.Spec = bpsv1alpha1.BestPracticeResultSpec{
				ScannerRef:       scannerCR.Name,
				CheckName:        check.Name,
				Category:         check.Category,
				Description:      check.Description,
				ComplianceStatus: bpsv1alpha1.ComplianceStatus(checkResult.ComplianceStatus),
				Reason:           checkResult.Reason,
				Remediation:      check.Remediation,
				CatalogURL:       catalogURL,
				Details:          apiDetails,
				Timestamp:        now,
			}
			return nil
		})

		if err != nil {
			logger.Error(err, "Failed to upsert result", "result", resultName)
		} else {
			logger.V(1).Info("Result upserted", "result", resultName, "operation", op)
		}
	}

	// Delete stale results
	if err := r.deleteStaleResults(ctx, &scannerCR, resultNames); err != nil {
		logger.Error(err, "Failed to clean up stale results")
	}

	// Re-fetch the scanner to avoid conflict errors from concurrent updates
	if err := r.Get(ctx, req.NamespacedName, &scannerCR); err != nil {
		return ctrl.Result{}, err
	}

	// Update status
	scannerCR.Status.Phase = bpsv1alpha1.PhaseCompleted
	scannerCR.Status.LastScanTime = &now
	scannerCR.Status.Summary = &summary

	// Calculate next scan time using already-parsed interval
	var requeueAfter time.Duration
	if scanInterval > 0 {
		requeueAfter = scanInterval
		nextScan := metav1.NewTime(now.Add(scanInterval))
		scannerCR.Status.NextScanTime = &nextScan
	}

	if err := r.Status().Update(ctx, &scannerCR); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Scan completed",
		"total", summary.Total,
		"compliant", summary.Compliant,
		"nonCompliant", summary.NonCompliant,
		"skipped", summary.Skipped,
	)

	if requeueAfter > 0 {
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	// One-shot scan completed — tear down the probe DaemonSet
	if r.OperatorNamespace != "" {
		if err := probe.DeleteDaemonSet(ctx, r.Client, r.OperatorNamespace); err != nil {
			logger.Error(err, "Failed to delete probe DaemonSet after one-shot scan")
		} else {
			logger.Info("Probe DaemonSet deleted after one-shot scan")
		}
	}

	return ctrl.Result{}, nil
}

func (r *ScannerReconciler) deleteStaleResults(ctx context.Context, scannerCR *bpsv1alpha1.BestPracticeScanner, currentNames map[string]bool) error {
	var resultList bpsv1alpha1.BestPracticeResultList
	// Use field selector to filter by scannerRef server-side for efficiency
	if err := r.List(ctx, &resultList,
		client.InNamespace(scannerCR.Namespace),
		client.MatchingFields{"spec.scannerRef": scannerCR.Name},
	); err != nil {
		return err
	}

	for i := range resultList.Items {
		result := &resultList.Items[i]
		if !currentNames[result.Name] {
			if err := r.Delete(ctx, result); err != nil && !errors.IsNotFound(err) {
				return err
			}
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScannerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Index BestPracticeResults by scannerRef for efficient querying
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &bpsv1alpha1.BestPracticeResult{}, "spec.scannerRef", func(obj client.Object) []string {
		result := obj.(*bpsv1alpha1.BestPracticeResult)
		return []string{result.Spec.ScannerRef}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&bpsv1alpha1.BestPracticeScanner{}).
		Owns(&bpsv1alpha1.BestPracticeResult{}).
		Named("scanner").
		Complete(r)
}
