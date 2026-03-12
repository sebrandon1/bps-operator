package performance

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// --- Memory limit ---

func TestCheckMemoryLimit_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "c1",
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("256Mi")},
					},
				}},
			},
		}},
	}
	result := CheckMemoryLimit(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckMemoryLimit_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "c1"}},
			},
		}},
	}
	result := CheckMemoryLimit(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckMemoryLimit_NoPods(t *testing.T) {
	result := CheckMemoryLimit(&checks.DiscoveredResources{})
	if result.ComplianceStatus != "Skipped" {
		t.Errorf("expected Skipped, got %s", result.ComplianceStatus)
	}
}

// --- Exclusive CPU pool ---

func TestCheckExclusiveCPUPool_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "c1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")},
						Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")},
					},
				}},
			},
		}},
	}
	result := CheckExclusiveCPUPool(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckExclusiveCPUPool_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "c1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")},
						Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("4")},
					},
				}},
			},
		}},
	}
	result := CheckExclusiveCPUPool(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckExclusiveCPUPool_FractionalCPU_Skipped(t *testing.T) {
	// Fractional CPU requests are not checked (not whole-CPU requests)
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "c1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")},
						Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
					},
				}},
			},
		}},
	}
	result := CheckExclusiveCPUPool(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant (fractional CPU not checked), got %s", result.ComplianceStatus)
	}
}

// --- RT apps no exec probes ---

func TestCheckRTAppsNoExecProbes_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns1",
				Annotations: map[string]string{"rt-app": "true"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "c1",
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							Exec: &corev1.ExecAction{Command: []string{"cat", "/tmp/healthy"}},
						},
					},
				}},
			},
		}},
	}
	result := CheckRTAppsNoExecProbes(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckRTAppsNoExecProbes_NonRT_Compliant(t *testing.T) {
	// Non-RT pods with exec probes are fine
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "c1",
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							Exec: &corev1.ExecAction{Command: []string{"cat", "/tmp/healthy"}},
						},
					},
				}},
			},
		}},
	}
	result := CheckRTAppsNoExecProbes(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant (non-RT pod), got %s", result.ComplianceStatus)
	}
}
