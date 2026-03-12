package networking

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// --- Dual-stack service ---

func TestCheckDualStackService_Compliant(t *testing.T) {
	policy := corev1.IPFamilyPolicyPreferDualStack
	resources := &checks.DiscoveredResources{
		Services: []corev1.Service{{
			ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "ns1"},
			Spec: corev1.ServiceSpec{
				IPFamilyPolicy: &policy,
			},
		}},
	}
	result := CheckDualStackService(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckDualStackService_NonCompliant(t *testing.T) {
	policy := corev1.IPFamilyPolicySingleStack
	resources := &checks.DiscoveredResources{
		Services: []corev1.Service{{
			ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "ns1"},
			Spec: corev1.ServiceSpec{
				IPFamilyPolicy: &policy,
			},
		}},
	}
	result := CheckDualStackService(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckDualStackService_Headless_Skipped(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Services: []corev1.Service{{
			ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "ns1"},
			Spec: corev1.ServiceSpec{
				ClusterIP: "None",
			},
		}},
	}
	result := CheckDualStackService(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant (headless skipped), got %s", result.ComplianceStatus)
	}
}

// --- NetworkPolicy deny-all ---

func TestCheckNetworkPolicyDenyAll_Compliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
		}},
		NetworkPolicies: []networkingv1.NetworkPolicy{{
			ObjectMeta: metav1.ObjectMeta{Name: "deny-all", Namespace: "ns1"},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []networkingv1.PolicyType{
					networkingv1.PolicyTypeIngress,
					networkingv1.PolicyTypeEgress,
				},
			},
		}},
		Namespaces: []string{"ns1"},
	}
	result := CheckNetworkPolicyDenyAll(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckNetworkPolicyDenyAll_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
		}},
		NetworkPolicies: []networkingv1.NetworkPolicy{},
		Namespaces:      []string{"ns1"},
	}
	result := CheckNetworkPolicyDenyAll(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

// --- Reserved ports ---

func TestCheckReservedPartnerPorts_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:  "c1",
					Ports: []corev1.ContainerPort{{ContainerPort: 22222}},
				}},
			},
		}},
	}
	result := CheckReservedPartnerPorts(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckReservedPartnerPorts_Compliant(t *testing.T) {
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
	result := CheckReservedPartnerPorts(resources)
	if result.ComplianceStatus != "Compliant" {
		t.Errorf("expected Compliant, got %s", result.ComplianceStatus)
	}
}

func TestCheckOCPReservedPorts_NonCompliant(t *testing.T) {
	resources := &checks.DiscoveredResources{
		Pods: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:  "c1",
					Ports: []corev1.ContainerPort{{ContainerPort: 22623}},
				}},
			},
		}},
	}
	result := CheckOCPReservedPorts(resources)
	if result.ComplianceStatus != "NonCompliant" {
		t.Errorf("expected NonCompliant, got %s", result.ComplianceStatus)
	}
}
