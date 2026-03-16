package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BestPracticeScannerSpec defines the desired state of BestPracticeScanner.
type BestPracticeScannerSpec struct {
	// TargetNamespace is the namespace to scan. Defaults to the CR's namespace if omitted.
	// +optional
	TargetNamespace string `json:"targetNamespace,omitempty"`

	// LabelSelector filters pods to scan.
	// +optional
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`

	// ScanInterval defines the interval between periodic scans (e.g. "5m"). Omit for one-shot.
	// +optional
	// +kubebuilder:validation:Pattern=`^([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$`
	ScanInterval string `json:"scanInterval,omitempty"`

	// Checks is an optional list of check names to run. Empty means run all.
	// +optional
	// +kubebuilder:validation:MinItems=1
	Checks []string `json:"checks,omitempty"`

	// Suspend pauses scanning when set to true.
	// +optional
	Suspend bool `json:"suspend,omitempty"`
}

// ScannerPhase represents the current phase of the scanner.
type ScannerPhase string

const (
	PhaseIdle      ScannerPhase = "Idle"
	PhaseScanning  ScannerPhase = "Scanning"
	PhaseCompleted ScannerPhase = "Completed"
	PhaseError     ScannerPhase = "Error"
)

// ScanSummary holds aggregate counts from a scan.
type ScanSummary struct {
	// Total number of checks run.
	Total int `json:"total"`
	// Number of compliant checks.
	Compliant int `json:"compliant"`
	// Number of non-compliant checks.
	NonCompliant int `json:"nonCompliant"`
	// Number of checks that errored.
	Error int `json:"error"`
	// Number of skipped checks.
	Skipped int `json:"skipped"`
}

// BestPracticeScannerStatus defines the observed state of BestPracticeScanner.
type BestPracticeScannerStatus struct {
	// Phase is the current phase of the scanner.
	// +optional
	Phase ScannerPhase `json:"phase,omitempty"`
	// LastScanTime is the timestamp of the last completed scan.
	// +optional
	LastScanTime *metav1.Time `json:"lastScanTime,omitempty"`
	// NextScanTime is the scheduled time for the next scan.
	// +optional
	NextScanTime *metav1.Time `json:"nextScanTime,omitempty"`
	// Summary holds aggregate results from the last scan.
	// +optional
	Summary *ScanSummary `json:"summary,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=bps
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Interval",type=string,JSONPath=`.spec.scanInterval`,priority=0
// +kubebuilder:printcolumn:name="Compliant",type=integer,JSONPath=`.status.summary.compliant`
// +kubebuilder:printcolumn:name="NonCompliant",type=integer,JSONPath=`.status.summary.nonCompliant`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// BestPracticeScanner is the Schema for the bestpracticescanners API.
type BestPracticeScanner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BestPracticeScannerSpec   `json:"spec,omitempty"`
	Status BestPracticeScannerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BestPracticeScannerList contains a list of BestPracticeScanner.
type BestPracticeScannerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BestPracticeScanner `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BestPracticeScanner{}, &BestPracticeScannerList{})
}
