package observability

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

func TestCheckCRDStatus_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		CRDs: []apiextv1.CustomResourceDefinition{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "tests.example.com"},
				Spec: apiextv1.CustomResourceDefinitionSpec{
					Versions: []apiextv1.CustomResourceDefinitionVersion{
						{
							Name: "v1",
							Subresources: &apiextv1.CustomResourceSubresources{
								Status: &apiextv1.CustomResourceSubresourceStatus{},
							},
						},
					},
				},
			},
		},
	}
	result := CheckCRDStatus(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckCRDStatus_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		CRDs: []apiextv1.CustomResourceDefinition{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "tests.example.com"},
				Spec: apiextv1.CustomResourceDefinitionSpec{
					Versions: []apiextv1.CustomResourceDefinitionVersion{
						{Name: "v1"},
					},
				},
			},
		},
	}
	result := CheckCRDStatus(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckTerminationPolicy_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1", TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError},
					},
				},
			},
		},
	}
	result := CheckTerminationPolicy(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckTerminationPolicy_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1"},
					},
				},
			},
		},
	}
	result := CheckTerminationPolicy(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckTerminationPolicy_NoPods(t *testing.T) {
	result := CheckTerminationPolicy(&checks.DiscoveredResources{})
	if result.ComplianceStatus != "Skipped" {
		t.Errorf("expected Skipped, got %s", result.ComplianceStatus)
	}
}
