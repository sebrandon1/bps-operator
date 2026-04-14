package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/redhat-best-practices-for-k8s/checks"
	bpsv1alpha1 "github.com/sebrandon1/bps-operator/api/v1alpha1"
	bpsmetrics "github.com/sebrandon1/bps-operator/internal/metrics"
	"github.com/sebrandon1/bps-operator/internal/probe"
	"github.com/sebrandon1/bps-operator/internal/scanner"
)

// ScannerReconciler reconciles a BestPracticeScanner object.
type ScannerReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	Recorder          record.EventRecorder
	ProbeExecutor     checks.ProbeExecutor
	OperatorNamespace string
	ProbeImage        string
	ScannerNodeName   string
	CertValidator     checks.CertificationValidator
	DiscoveryClient   discovery.ServerVersionInterface
	K8sClientset      any // kubernetes.Interface
	ScaleClient       any // scale.ScalesGetter
	CatalogURLBase    string
}

const probeRequeueInterval = 5 * time.Second

var errProbesPending = fmt.Errorf("probe pods not yet running")

// +kubebuilder:rbac:groups=bps.openshift.io,resources=bestpracticescanners,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bps.openshift.io,resources=bestpracticescanners/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=bps.openshift.io,resources=bestpracticescanners/finalizers,verbs=update
// +kubebuilder:rbac:groups=bps.openshift.io,resources=bestpracticeresults,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods;services;serviceaccounts;namespaces;nodes;persistentvolumes;resourcequotas;secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods,verbs=delete
// +kubebuilder:rbac:groups="",resources=nodes,verbs=patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings;clusterrolebindings,verbs=get;list;watch
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=daemonsets;deployments;statefulsets,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=pods/exec,verbs=create
// +kubebuilder:rbac:groups="",resources=pods/log,verbs=get
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch
// +kubebuilder:rbac:groups=storage.k8s.io,resources=storageclasses,verbs=get;list;watch
// +kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions;clusteroperators,verbs=get;list;watch
// +kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions;catalogsources;subscriptions,verbs=get;list;watch
// +kubebuilder:rbac:groups=packages.operators.coreos.com,resources=packagemanifests,verbs=get;list;watch
// +kubebuilder:rbac:groups=apiserver.openshift.io,resources=apirequestcounts,verbs=get;list;watch
// +kubebuilder:rbac:groups=k8s.cni.cncf.io,resources=network-attachment-definitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=sriovnetwork.openshift.io,resources=sriovnetworks;sriovnetworknodepolicies,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *ScannerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the BestPracticeScanner CR
	var scannerCR bpsv1alpha1.BestPracticeScanner
	if err := r.Get(ctx, req.NamespacedName, &scannerCR); err != nil {
		if apierrors.IsNotFound(err) {
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
			return ctrl.Result{}, nil
		}
		if time.Since(scannerCR.Status.LastScanTime.Time) < scanInterval {
			requeueIn := time.Until(scannerCR.Status.LastScanTime.Add(scanInterval))
			return ctrl.Result{RequeueAfter: requeueIn}, nil
		}
	}

	// Enforce one scanner per namespace
	if err := r.enforceUniqueness(ctx, &scannerCR); err != nil {
		return ctrl.Result{}, err
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
	meta.SetStatusCondition(&scannerCR.Status.Conditions, metav1.Condition{
		Type:               bpsv1alpha1.ConditionScanComplete,
		Status:             metav1.ConditionFalse,
		Reason:             bpsv1alpha1.ReasonScanStarted,
		Message:            "Compliance scan in progress",
		ObservedGeneration: scannerCR.Generation,
	})
	if err := r.Status().Update(ctx, &scannerCR); err != nil {
		return ctrl.Result{}, err
	}
	r.Recorder.Event(&scannerCR, corev1.EventTypeNormal, bpsv1alpha1.ReasonScanStarted, "Starting compliance scan")

	// Discover and run checks
	scanStart := time.Now()

	resources, err := r.discoverResources(ctx, &scannerCR, targetNS)
	if errors.Is(err, errProbesPending) {
		logger.Info("Probe pods not yet running, requeueing")
		return ctrl.Result{RequeueAfter: probeRequeueInterval}, nil
	}
	if err != nil {
		r.Recorder.Eventf(&scannerCR, corev1.EventTypeWarning, bpsv1alpha1.ReasonScanFailed, "Resource discovery failed: %v", err)
		return ctrl.Result{}, err
	}

	summary, resultNames := r.runChecks(ctx, &scannerCR, resources)

	scanDuration := time.Since(scanStart)

	// Delete stale results
	if err := r.deleteStaleResults(ctx, &scannerCR, resultNames); err != nil {
		logger.Error(err, "Failed to clean up stale results")
	}

	return r.completeScan(ctx, req, &scannerCR, summary, scanInterval, scanDuration)
}

// enforceUniqueness ensures only one scanner exists per namespace.
func (r *ScannerReconciler) enforceUniqueness(ctx context.Context, scannerCR *bpsv1alpha1.BestPracticeScanner) error {
	logger := log.FromContext(ctx)

	var scannerList bpsv1alpha1.BestPracticeScannerList
	if err := r.List(ctx, &scannerList, client.InNamespace(scannerCR.Namespace)); err != nil {
		return err
	}
	for i := range scannerList.Items {
		other := &scannerList.Items[i]
		if other.Name != scannerCR.Name && other.CreationTimestamp.Before(&scannerCR.CreationTimestamp) {
			logger.Info("Another scanner already exists in this namespace, setting phase to Error",
				"existing", other.Name)
			scannerCR.Status.Phase = bpsv1alpha1.PhaseError
			if err := r.Status().Update(ctx, scannerCR); err != nil {
				logger.Error(err, "Failed to update scanner status to Error")
			}
			return fmt.Errorf("namespace %s already has scanner %q; only one scanner per namespace is allowed", scannerCR.Namespace, other.Name)
		}
	}
	return nil
}

func (r *ScannerReconciler) discoverResources(ctx context.Context, scannerCR *bpsv1alpha1.BestPracticeScanner, targetNS string) (*checks.DiscoveredResources, error) {
	logger := log.FromContext(ctx)

	// Ensure probe DaemonSet
	if r.OperatorNamespace != "" {
		if err := probe.EnsureDaemonSet(ctx, r.Client, r.OperatorNamespace, r.ProbeImage); err != nil {
			logger.Error(err, "Failed to ensure probe DaemonSet")
		}
	}

	resources, err := scanner.Discover(ctx, r.Client, targetNS, scannerCR.Spec.LabelSelector, r.DiscoveryClient)
	if err != nil {
		scannerCR.Status.Phase = bpsv1alpha1.PhaseError
		meta.SetStatusCondition(&scannerCR.Status.Conditions, metav1.Condition{
			Type:               bpsv1alpha1.ConditionScanComplete,
			Status:             metav1.ConditionFalse,
			Reason:             bpsv1alpha1.ReasonScanFailed,
			Message:            fmt.Sprintf("Resource discovery failed: %v", err),
			ObservedGeneration: scannerCR.Generation,
		})
		if updateErr := r.Status().Update(ctx, scannerCR); updateErr != nil {
			logger.Error(updateErr, "Failed to update scanner status to Error")
		}
		return nil, fmt.Errorf("discovering resources: %w", err)
	}

	// Wire fields not populated by Discover
	resources.ScannerPodNodeName = r.ScannerNodeName
	resources.CertValidator = r.CertValidator
	resources.K8sClientset = r.K8sClientset
	resources.ScaleClient = r.ScaleClient

	if r.OperatorNamespace != "" {
		probePods, err := probe.MapProbePods(ctx, r.Client, r.OperatorNamespace)
		if err != nil {
			logger.Error(err, "Failed to map probe pods, probe-based checks will be skipped")
			r.Recorder.Event(scannerCR, corev1.EventTypeWarning, bpsv1alpha1.ReasonProbeUnavailable, "Probe pods not available, probe-based checks will be skipped")
			meta.SetStatusCondition(&scannerCR.Status.Conditions, metav1.Condition{
				Type:               bpsv1alpha1.ConditionProbeAvailable,
				Status:             metav1.ConditionFalse,
				Reason:             bpsv1alpha1.ReasonProbePodsFailed,
				Message:            err.Error(),
				ObservedGeneration: scannerCR.Generation,
			})
		} else if len(probePods) == 0 {
			return nil, errProbesPending
		} else {
			resources.ProbePods = probePods
			resources.ProbeExecutor = r.ProbeExecutor
			meta.SetStatusCondition(&scannerCR.Status.Conditions, metav1.Condition{
				Type:               bpsv1alpha1.ConditionProbeAvailable,
				Status:             metav1.ConditionTrue,
				Reason:             bpsv1alpha1.ReasonProbePodsReady,
				Message:            fmt.Sprintf("%d probe pods available", len(probePods)),
				ObservedGeneration: scannerCR.Generation,
			})
		}
	}

	return resources, nil
}

// runChecks executes all applicable checks and upserts BestPracticeResult objects.
// Returns the scan summary and the set of current result names.
func (r *ScannerReconciler) runChecks(ctx context.Context, scannerCR *bpsv1alpha1.BestPracticeScanner, resources *checks.DiscoveredResources) (bpsv1alpha1.ScanSummary, map[string]bool) {
	checksToRun := checks.Filtered(scannerCR.Spec.Checks)
	now := metav1.Now()
	summary := bpsv1alpha1.ScanSummary{Total: len(checksToRun)}
	resultNames := make(map[string]bool)

	for _, check := range checksToRun {
		checkResult := check.Fn(resources)

		switch bpsv1alpha1.ComplianceStatus(checkResult.ComplianceStatus) {
		case bpsv1alpha1.StatusCompliant:
			summary.Compliant++
		case bpsv1alpha1.StatusNonCompliant:
			summary.NonCompliant++
		case bpsv1alpha1.StatusError:
			summary.Error++
		case bpsv1alpha1.StatusSkipped:
			summary.Skipped++
		}

		resultName := fmt.Sprintf("%s-%s", scannerCR.Name, check.Name)
		resultNames[resultName] = true

		r.upsertResult(ctx, scannerCR, check, checkResult, resultName, now)
	}

	return summary, resultNames
}

// upsertResult creates or updates a BestPracticeResult for a single check.
func (r *ScannerReconciler) upsertResult(ctx context.Context, scannerCR *bpsv1alpha1.BestPracticeScanner, check checks.CheckInfo, checkResult checks.CheckResult, resultName string, now metav1.Time) {
	logger := log.FromContext(ctx)

	result := &bpsv1alpha1.BestPracticeResult{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resultName,
			Namespace: scannerCR.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, result, func() error {
		if err := controllerutil.SetControllerReference(scannerCR, result, r.Scheme); err != nil {
			return err
		}

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
		if check.CatalogID != "" && r.CatalogURLBase != "" {
			catalogURL = r.CatalogURLBase + "#" + check.CatalogID
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

// completeScan finalizes the scan by updating status, recording metrics, and handling periodic requeue/probe cleanup.
func (r *ScannerReconciler) completeScan(ctx context.Context, req ctrl.Request, scannerCR *bpsv1alpha1.BestPracticeScanner, summary bpsv1alpha1.ScanSummary, scanInterval, scanDuration time.Duration) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Record metrics
	name, ns := scannerCR.Name, scannerCR.Namespace
	bpsmetrics.ScanDuration.WithLabelValues(name, ns).Observe(scanDuration.Seconds())
	bpsmetrics.ScanTotal.WithLabelValues(name, ns).Inc()
	bpsmetrics.CheckResults.WithLabelValues(name, ns, "Compliant").Set(float64(summary.Compliant))
	bpsmetrics.CheckResults.WithLabelValues(name, ns, "NonCompliant").Set(float64(summary.NonCompliant))
	bpsmetrics.CheckResults.WithLabelValues(name, ns, "Skipped").Set(float64(summary.Skipped))
	bpsmetrics.CheckResults.WithLabelValues(name, ns, "Error").Set(float64(summary.Error))

	// Re-fetch the scanner to avoid conflict errors from concurrent updates
	if err := r.Get(ctx, req.NamespacedName, scannerCR); err != nil {
		return ctrl.Result{}, err
	}

	now := metav1.Now()
	scannerCR.Status.Phase = bpsv1alpha1.PhaseCompleted
	scannerCR.Status.LastScanTime = &now
	scannerCR.Status.Summary = &summary

	meta.SetStatusCondition(&scannerCR.Status.Conditions, metav1.Condition{
		Type:               bpsv1alpha1.ConditionScanComplete,
		Status:             metav1.ConditionTrue,
		Reason:             bpsv1alpha1.ReasonScanSucceeded,
		Message:            fmt.Sprintf("%d checks: %d compliant, %d non-compliant, %d skipped, %d errors", summary.Total, summary.Compliant, summary.NonCompliant, summary.Skipped, summary.Error),
		ObservedGeneration: scannerCR.Generation,
	})

	var requeueAfter time.Duration
	if scanInterval > 0 {
		requeueAfter = scanInterval
		nextScan := metav1.NewTime(now.Add(scanInterval))
		scannerCR.Status.NextScanTime = &nextScan
	}

	if err := r.Status().Update(ctx, scannerCR); err != nil {
		return ctrl.Result{}, err
	}

	r.Recorder.Eventf(scannerCR, corev1.EventTypeNormal, bpsv1alpha1.ReasonScanCompleted,
		"Scan completed: %d compliant, %d non-compliant, %d skipped, %d errors (%.1fs)",
		summary.Compliant, summary.NonCompliant, summary.Skipped, summary.Error, scanDuration.Seconds())

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
	if err := r.List(ctx, &resultList,
		client.InNamespace(scannerCR.Namespace),
		client.MatchingFields{"spec.scannerRef": scannerCR.Name},
	); err != nil {
		return err
	}

	for i := range resultList.Items {
		result := &resultList.Items[i]
		if !currentNames[result.Name] {
			if err := r.Delete(ctx, result); err != nil && !apierrors.IsNotFound(err) {
				return err
			}
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScannerReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
