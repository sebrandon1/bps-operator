package webhook

import (
	"context"
	"strings"
	"testing"

	"github.com/redhat-best-practices-for-k8s/checks"
	checksall "github.com/redhat-best-practices-for-k8s/checks/all"
	bpsv1alpha1 "github.com/redhat-best-practices-for-k8s/checks-types/api/v1alpha1"
)

// knownCheckName returns a real check name from the registered catalog and
// tolerates repeated calls: registration itself is guarded by sync.Once.
func knownCheckName(t *testing.T) string {
	t.Helper()
	registerChecks.Do(checksall.Register)
	all := checks.All()
	if len(all) == 0 {
		t.Fatal("no checks registered")
	}
	return all[0].Name
}

func TestValidate(t *testing.T) {
	known := knownCheckName(t)

	tests := []struct {
		name     string
		spec     bpsv1alpha1.BestPracticeScannerSpec
		wantErrs []string
	}{
		{name: "empty spec"},
		{name: "nil checks", spec: bpsv1alpha1.BestPracticeScannerSpec{ScanInterval: "5m"}},
		{name: "valid interval", spec: bpsv1alpha1.BestPracticeScannerSpec{ScanInterval: "1h30m"}},
		{name: "invalid interval", spec: bpsv1alpha1.BestPracticeScannerSpec{ScanInterval: "bogus"}, wantErrs: []string{"is not a valid duration"}},
		{name: "known check", spec: bpsv1alpha1.BestPracticeScannerSpec{Checks: []string{known}}},
		{
			name:     "unknown check",
			spec:     bpsv1alpha1.BestPracticeScannerSpec{Checks: []string{known, "not-a-real-check"}},
			wantErrs: []string{"unknown check name", "make list-checks"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &BestPracticeScannerValidator{}
			scanner := &bpsv1alpha1.BestPracticeScanner{Spec: tt.spec}

			_, createErr := v.ValidateCreate(context.Background(), scanner)
			_, updateErr := v.ValidateUpdate(context.Background(), scanner, scanner)
			for _, err := range []error{createErr, updateErr} {
				if len(tt.wantErrs) == 0 {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
					continue
				}
				if err == nil {
					t.Fatalf("error = nil, want containing %q", tt.wantErrs)
				}
				for _, want := range tt.wantErrs {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error = %v, want containing %q", err, want)
					}
				}
			}
		})
	}
}

func TestValidateChecks_RepeatedCalls(t *testing.T) {
	known := knownCheckName(t)
	for i := range 3 {
		if err := validateChecks([]string{known}); err != nil {
			t.Fatalf("call %d: unexpected error for known check: %v", i+1, err)
		}
	}
}

func TestValidateDelete(t *testing.T) {
	v := &BestPracticeScannerValidator{}
	warnings, err := v.ValidateDelete(context.Background(), &bpsv1alpha1.BestPracticeScanner{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if warnings != nil {
		t.Errorf("warnings = %v, want nil", warnings)
	}
}
