package accesscontrol

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

func boolPtr(b bool) *bool       { return &b }
func int64Ptr(i int64) *int64    { return &i }

func TestCheckHostNetwork_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"}},
		},
	}
	result := CheckHostNetwork(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckHostNetwork_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec:       corev1.PodSpec{HostNetwork: true},
			},
		},
	}
	result := CheckHostNetwork(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
	if len(result.Details) != 1 {
		t.Errorf("expected 1 detail, got %d", len(result.Details))
	}
}

func TestCheckHostNetwork_NoPods(t *testing.T) {
	result := CheckHostNetwork(&checks.DiscoveredResources{})
	if result.ComplianceStatus != "Skipped" {
		t.Errorf("expected Skipped, got %s", result.ComplianceStatus)
	}
}

func TestCheckHostPath_NonCompliant(t *testing.T) {
	hostPathType := corev1.HostPathDirectory
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: "host-vol",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{Path: "/data", Type: &hostPathType},
						},
					}},
				},
			},
		},
	}
	result := CheckHostPath(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckHostIPC_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec:       corev1.PodSpec{HostIPC: true},
			},
		},
	}
	result := CheckHostIPC(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckHostPID_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec:       corev1.PodSpec{HostPID: true},
			},
		},
	}
	result := CheckHostPID(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckContainerHostPort_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1", Ports: []corev1.ContainerPort{{HostPort: 8080}}},
					},
				},
			},
		},
	}
	result := CheckContainerHostPort(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckSysAdmin_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "c1",
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{"SYS_ADMIN"},
								},
							},
						},
					},
				},
			},
		},
	}
	result := CheckSysAdmin(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckSysAdmin_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "c1"}},
				},
			},
		},
	}
	result := CheckSysAdmin(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckNonRootUser_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "c1"}},
				},
			},
		},
	}
	result := CheckNonRootUser(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckNonRootUser_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1", SecurityContext: &corev1.SecurityContext{RunAsNonRoot: boolPtr(true)}},
					},
				},
			},
		},
	}
	result := CheckNonRootUser(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckNonRootUser_PodLevel(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{RunAsNonRoot: boolPtr(true)},
					Containers:      []corev1.Container{{Name: "c1"}},
				},
			},
		},
	}
	result := CheckNonRootUser(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckPrivilegeEscalation_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "c1"}},
				},
			},
		},
	}
	result := CheckPrivilegeEscalation(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckPrivilegeEscalation_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1", SecurityContext: &corev1.SecurityContext{AllowPrivilegeEscalation: boolPtr(false)}},
					},
				},
			},
		},
	}
	result := CheckPrivilegeEscalation(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckReadOnlyFilesystem_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "c1"}},
				},
			},
		},
	}
	result := CheckReadOnlyFilesystem(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheck1337UID_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1", SecurityContext: &corev1.SecurityContext{RunAsUser: int64Ptr(1337)}},
					},
				},
			},
		},
	}
	result := Check1337UID(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheck1337UID_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1", SecurityContext: &corev1.SecurityContext{RunAsUser: int64Ptr(1000)}},
					},
				},
			},
		},
	}
	result := Check1337UID(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckServiceAccount_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec:       corev1.PodSpec{ServiceAccountName: "default"},
			},
		},
	}
	result := CheckServiceAccount(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckServiceAccount_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec:       corev1.PodSpec{ServiceAccountName: "my-sa"},
			},
		},
	}
	result := CheckServiceAccount(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckNodePortService_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Services: []corev1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "ns1"},
				Spec:       corev1.ServiceSpec{Type: corev1.ServiceTypeNodePort},
			},
		},
	}
	result := CheckNodePortService(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckNodePortService_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Services: []corev1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "ns1"},
				Spec:       corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP},
			},
		},
	}
	result := CheckNodePortService(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckAutomountToken_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec:       corev1.PodSpec{ServiceAccountName: "my-sa"},
			},
		},
		ServiceAccounts: []corev1.ServiceAccount{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "my-sa", Namespace: "ns1"},
				// AutomountServiceAccountToken not set → defaults to true
			},
		},
	}
	result := CheckAutomountToken(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckAutomountToken_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					ServiceAccountName:           "my-sa",
					AutomountServiceAccountToken: boolPtr(false),
				},
			},
		},
	}
	result := CheckAutomountToken(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckClusterRoleBindings_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec:       corev1.PodSpec{ServiceAccountName: "my-sa"},
			},
		},
		ClusterRoleBindings: []rbacv1.ClusterRoleBinding{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "crb1"},
				Subjects: []rbacv1.Subject{
					{Kind: "ServiceAccount", Name: "my-sa", Namespace: "ns1"},
				},
			},
		},
	}
	result := CheckClusterRoleBindings(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckSecurityContext_Privileged(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1", SecurityContext: &corev1.SecurityContext{Privileged: boolPtr(true)}},
					},
				},
			},
		},
	}
	result := CheckSecurityContext(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckCapabilities_AllTypes(t *testing.T) {
	tests := []struct {
		name    string
		cap     string
		checkFn checks.CheckFunc
	}{
		{"NET_ADMIN", "NET_ADMIN", CheckNetAdmin},
		{"NET_RAW", "NET_RAW", CheckNetRaw},
		{"IPC_LOCK", "IPC_LOCK", CheckIPCLock},
		{"BPF", "BPF", CheckBPF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := &checks.DiscoveredResources{
				Pods: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "c1",
									SecurityContext: &corev1.SecurityContext{
										Capabilities: &corev1.Capabilities{
											Add: []corev1.Capability{corev1.Capability(tt.cap)},
										},
									},
								},
							},
						},
					},
				},
			}
			result := tt.checkFn(resources)
			if result.ComplianceStatus != "NonCompliant" {
				t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
			}
		})
	}
}
