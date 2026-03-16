package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComplianceStatus represents the compliance state of a check.
type ComplianceStatus string

const (
	StatusCompliant    ComplianceStatus = "Compliant"
	StatusNonCompliant ComplianceStatus = "NonCompliant"
	StatusError        ComplianceStatus = "Error"
	StatusSkipped      ComplianceStatus = "Skipped"
)

// ResourceDetail describes a specific resource's compliance status.
type ResourceDetail struct {
	// Kind of the Kubernetes resource.
	Kind string `json:"kind"`
	// Name of the resource.
	Name string `json:"name"`
	// Namespace of the resource.
	Namespace string `json:"namespace"`
	// Compliant indicates whether this specific resource is compliant.
	Compliant bool `json:"compliant"`
	// Message provides details about the compliance status.
	Message string `json:"message"`
}

// BestPracticeResultSpec defines the observed result of a best practice check.
type BestPracticeResultSpec struct {
	// ScannerRef is the name of the BestPracticeScanner that produced this result.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	ScannerRef string `json:"scannerRef"`
	// CheckName is the unique identifier for the check.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	CheckName string `json:"checkName"`
	// Category groups the check (e.g. "access-control", "observability").
	Category string `json:"category"`
	// Description explains what the check verifies.
	Description string `json:"description"`
	// ComplianceStatus is the result of the check.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Compliant;NonCompliant;Error;Skipped
	ComplianceStatus ComplianceStatus `json:"complianceStatus"`
	// Reason explains why the check has its current status.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Remediation suggests how to fix non-compliance.
	// +optional
	Remediation string `json:"remediation,omitempty"`
	// CatalogURL links to the certsuite CATALOG.md entry for this check.
	// +optional
	CatalogURL string `json:"catalogURL,omitempty"`
	// Details lists per-resource compliance information.
	// +optional
	Details []ResourceDetail `json:"details,omitempty"`
	// Timestamp is when the check was executed.
	Timestamp metav1.Time `json:"timestamp"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=bpr
// +kubebuilder:printcolumn:name="Check",type=string,JSONPath=`.spec.checkName`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.spec.complianceStatus`
// +kubebuilder:printcolumn:name="Scanner",type=string,JSONPath=`.spec.scannerRef`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// BestPracticeResult is the Schema for the bestpracticeresults API.
type BestPracticeResult struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BestPracticeResultSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// BestPracticeResultList contains a list of BestPracticeResult.
type BestPracticeResultList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BestPracticeResult `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BestPracticeResult{}, &BestPracticeResultList{})
}
