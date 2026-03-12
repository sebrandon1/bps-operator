package lifecycle

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

func int32Ptr(i int32) *int32 { return &i }

// --- Probe checks ---

func TestCheckStartupProbe_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:         "c1",
					StartupProbe: &corev1.Probe{},
				}},
			},
		}},
	}
	result := CheckStartupProbe(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckStartupProbe_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "c1"}},
			},
		}},
	}
	result := CheckStartupProbe(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckReadinessProbe_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "c1"}},
			},
		}},
	}
	result := CheckReadinessProbe(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckLivenessProbe_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "c1"}},
			},
		}},
	}
	result := CheckLivenessProbe(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckProbes_NoPods(t *testing.T) {
	resources := &checks.DiscoveredResources{}
	for _, fn := range []checks.CheckFunc{CheckStartupProbe, CheckReadinessProbe, CheckLivenessProbe} {
		result := fn(resources)
		if result.ComplianceStatus != "Skipped" {
			t.Errorf("expected Skipped, got %s", result.ComplianceStatus)
		}
	}
}

// --- Hook checks ---

func TestCheckPreStop_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "c1"}},
			},
		}},
	}
	result := CheckPreStop(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckPreStop_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "c1",
					Lifecycle: &corev1.Lifecycle{
						PreStop: &corev1.LifecycleHandler{
							Exec: &corev1.ExecAction{Command: []string{"/bin/sh", "-c", "sleep 5"}},
						},
					},
				}},
			},
		}},
	}
	result := CheckPreStop(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckPostStart_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "c1"}},
			},
		}},
	}
	result := CheckPostStart(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

// --- Pod checks ---

func TestCheckImagePullPolicy_Always_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:            "c1",
					Image:           "nginx:latest",
					ImagePullPolicy: corev1.PullAlways,
				}},
			},
		}},
	}
	result := CheckImagePullPolicy(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckImagePullPolicy_IfNotPresent_WithDigest_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:            "c1",
					Image:           "nginx@sha256:abc123",
					ImagePullPolicy: corev1.PullIfNotPresent,
				}},
			},
		}},
	}
	result := CheckImagePullPolicy(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckImagePullPolicy_IfNotPresent_NoDigest_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:            "c1",
					Image:           "nginx:1.21",
					ImagePullPolicy: corev1.PullIfNotPresent,
				}},
			},
		}},
	}
	result := CheckImagePullPolicy(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckPodOwnerType_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns1",
				OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet", Name: "rs1"}},
			},
		}},
	}
	result := CheckPodOwnerType(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckPodOwnerType_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
		}},
	}
	result := CheckPodOwnerType(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckHighAvailability_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Deployments: []appsv1.Deployment{{
			ObjectMeta: metav1.ObjectMeta{Name: "deploy1", Namespace: "ns1"},
			Spec:       appsv1.DeploymentSpec{Replicas: int32Ptr(3)},
		}},
	}
	result := CheckHighAvailability(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckHighAvailability_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Deployments: []appsv1.Deployment{{
			ObjectMeta: metav1.ObjectMeta{Name: "deploy1", Namespace: "ns1"},
			Spec:       appsv1.DeploymentSpec{Replicas: int32Ptr(1)},
		}},
	}
	result := CheckHighAvailability(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckHighAvailability_NilReplicas(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Deployments: []appsv1.Deployment{{
			ObjectMeta: metav1.ObjectMeta{Name: "deploy1", Namespace: "ns1"},
			Spec:       appsv1.DeploymentSpec{},
		}},
	}
	result := CheckHighAvailability(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant (nil replicas defaults to 1), got %s", result.ComplianceStatus)
	}
}

func TestCheckCPUIsolation_Compliant(t *testing.T) {
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
	result := CheckCPUIsolation(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckCPUIsolation_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "c1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
						Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")},
					},
				}},
			},
		}},
	}
	result := CheckCPUIsolation(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckTolerationBypass_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Tolerations: []corev1.Toleration{{
					Key:    "node-role.kubernetes.io/master",
					Effect: corev1.TaintEffectNoSchedule,
				}},
			},
		}},
	}
	result := CheckTolerationBypass(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckTolerationBypass_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Tolerations: []corev1.Toleration{{
					Key:    "some-other-taint",
					Effect: corev1.TaintEffectNoSchedule,
				}},
			},
		}},
	}
	result := CheckTolerationBypass(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckPVReclaimPolicy_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		PersistentVolumes: []corev1.PersistentVolume{{
			ObjectMeta: metav1.ObjectMeta{Name: "pv1"},
			Spec: corev1.PersistentVolumeSpec{
				PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			},
		}},
	}
	result := CheckPVReclaimPolicy(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckPVReclaimPolicy_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		PersistentVolumes: []corev1.PersistentVolume{{
			ObjectMeta: metav1.ObjectMeta{Name: "pv1"},
			Spec: corev1.PersistentVolumeSpec{
				PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			},
		}},
	}
	result := CheckPVReclaimPolicy(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckPodScheduling_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec:       corev1.PodSpec{},
		}},
	}
	result := CheckPodScheduling(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckPodScheduling_Compliant_NodeSelector(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				NodeSelector: map[string]string{"role": "worker"},
			},
		}},
	}
	result := CheckPodScheduling(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckAffinityRequired_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec:       corev1.PodSpec{},
		}},
	}
	result := CheckAffinityRequired(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}
