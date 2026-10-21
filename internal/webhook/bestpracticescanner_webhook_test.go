package webhook

import (
	"context"
	"strings"
	"testing"

	bpsv1alpha1 "github.com/redhat-best-practices-for-k8s/checks-types/api/v1alpha1"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    bpsv1alpha1.BestPracticeScannerSpec
		wantErr string
	}{
		{name: "empty spec"},
		{name: "valid interval", spec: bpsv1alpha1.BestPracticeScannerSpec{ScanInterval: "1h30m"}},
		{name: "invalid interval", spec: bpsv1alpha1.BestPracticeScannerSpec{ScanInterval: "5mins"}, wantErr: "is not a valid duration"},
		{name: "unknown check", spec: bpsv1alpha1.BestPracticeScannerSpec{Checks: []string{"not-a-real-check"}}, wantErr: "unknown check name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &BestPracticeScannerValidator{}
			scanner := &bpsv1alpha1.BestPracticeScanner{Spec: tt.spec}

			_, createErr := v.ValidateCreate(context.Background(), scanner)
			_, updateErr := v.ValidateUpdate(context.Background(), scanner, scanner)
			for _, err := range []error{createErr, updateErr} {
				if tt.wantErr == "" {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
				} else if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %v, want containing %q", err, tt.wantErr)
				}
			}
		})
	}
}

func TestValidateChecks_KnownCheck(t *testing.T) {
	if err := validateChecks([]string{"access-control-pod-host-network"}); err != nil {
		t.Errorf("unexpected error for known check: %v", err)
	}
	// Repeated calls must stay safe now that registration is guarded by sync.Once.
	if err := validateChecks([]string{"access-control-pod-host-network"}); err != nil {
		t.Errorf("unexpected error on second call: %v", err)
	}
}

func TestValidateDelete(t *testing.T) {
	v := &BestPracticeScannerValidator{}
	if _, err := v.ValidateDelete(context.Background(), &bpsv1alpha1.BestPracticeScanner{}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
