package manageability

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// --- Port name format ---

func TestCheckPortNameFormat_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:  "c1",
					Ports: []corev1.ContainerPort{{Name: "http", ContainerPort: 8080}},
				}},
			},
		}},
	}
	result := CheckPortNameFormat(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckPortNameFormat_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:  "c1",
					Ports: []corev1.ContainerPort{{Name: "HTTP_PORT", ContainerPort: 8080}},
				}},
			},
		}},
	}
	result := CheckPortNameFormat(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckPortNameFormat_UnnamedPort(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:  "c1",
					Ports: []corev1.ContainerPort{{ContainerPort: 8080}},
				}},
			},
		}},
	}
	result := CheckPortNameFormat(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant (unnamed ports not checked), got %s", result.ComplianceStatus)
	}
}

// --- Image tag ---

func TestCheckImageTag_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:  "c1",
					Image: "nginx:1.21",
				}},
			},
		}},
	}
	result := CheckImageTag(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckImageTag_Digest_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:  "c1",
					Image: "nginx@sha256:abc123def456",
				}},
			},
		}},
	}
	result := CheckImageTag(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckImageTag_Latest_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:  "c1",
					Image: "nginx:latest",
				}},
			},
		}},
	}
	result := CheckImageTag(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckImageTag_NoTag_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:  "c1",
					Image: "nginx",
				}},
			},
		}},
	}
	result := CheckImageTag(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}
