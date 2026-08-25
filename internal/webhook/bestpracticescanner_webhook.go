package webhook

import (
	"context"
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/redhat-best-practices-for-k8s/checks"
	checksall "github.com/redhat-best-practices-for-k8s/checks/all"
	bpsv1alpha1 "github.com/redhat-best-practices-for-k8s/checks-types/api/v1alpha1"
)

// +kubebuilder:webhook:path=/validate-bps-openshift-io-v1alpha1-bestpracticescanner,mutating=false,failurePolicy=fail,sideEffects=None,groups=bps.openshift.io,resources=bestpracticescanners,verbs=create;update,versions=v1alpha1,name=vbestpracticescanner.kb.io,admissionReviewVersions=v1

// BestPracticeScannerValidator validates BestPracticeScanner admission requests.
type BestPracticeScannerValidator struct{}

// SetupWithManager registers the validating webhook with the manager.
func (v *BestPracticeScannerValidator) SetupWithManager(mgr ctrl.Manager) error {
	return builder.WebhookManagedBy[*bpsv1alpha1.BestPracticeScanner](mgr, &bpsv1alpha1.BestPracticeScanner{}).
		WithValidator(v).
		Complete()
}

// ValidateCreate validates a new BestPracticeScanner.
func (v *BestPracticeScannerValidator) ValidateCreate(_ context.Context, scanner *bpsv1alpha1.BestPracticeScanner) (admission.Warnings, error) {
	return nil, validate(scanner)
}

// ValidateUpdate validates an updated BestPracticeScanner.
func (v *BestPracticeScannerValidator) ValidateUpdate(_ context.Context, _, scanner *bpsv1alpha1.BestPracticeScanner) (admission.Warnings, error) {
	return nil, validate(scanner)
}

// ValidateDelete always allows deletion.
func (v *BestPracticeScannerValidator) ValidateDelete(_ context.Context, _ *bpsv1alpha1.BestPracticeScanner) (admission.Warnings, error) {
	return nil, nil
}

func validate(scanner *bpsv1alpha1.BestPracticeScanner) error {
	if err := validateScanInterval(scanner.Spec.ScanInterval); err != nil {
		return err
	}
	return validateChecks(scanner.Spec.Checks)
}

func validateScanInterval(interval string) error {
	if interval == "" {
		return nil
	}
	if _, err := time.ParseDuration(interval); err != nil {
		return fmt.Errorf("spec.scanInterval %q is not a valid duration (e.g. 30m, 1h): %w", interval, err)
	}
	return nil
}

func validateChecks(checkNames []string) error {
	if len(checkNames) == 0 {
		return nil
	}
	checksall.Register()
	known := make(map[string]bool, len(checks.All()))
	for _, c := range checks.All() {
		known[c.Name] = true
	}
	for _, name := range checkNames {
		if !known[name] {
			return fmt.Errorf("spec.checks contains unknown check name %q; run 'make list-checks' to see valid names", name)
		}
	}
	return nil
}
