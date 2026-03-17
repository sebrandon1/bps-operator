// Package probe manages the certsuite-probe DaemonSet, which provides node-level
// access for compliance checks that require inspection of host configurations.
//
// # Security Model
//
// The probe runs with elevated privileges (privileged container, host namespaces,
// host root filesystem mount) to enable checks that verify node-level configurations.
// This is required for checks such as:
//   - Kernel parameter verification (sysctl settings)
//   - Host file inspection (/etc, /proc, /sys)
//   - Network configuration validation
//   - Container runtime inspection
//   - Security context validation
//
// # Security Boundaries
//
//  1. Namespace Isolation: Pods run in the operator namespace only, not in user namespaces
//     being scanned. This prevents privilege escalation in user workload namespaces.
//
//  2. Read-Only Host Mount: The host root filesystem is mounted read-only at /host,
//     preventing any modifications to the node.
//
//  3. No Automated Execution: The container runs "sleep infinity" with no automated
//     code execution. Commands are executed only via explicit Kubernetes RBAC-controlled
//     pods/exec API calls.
//
//  4. Execution Timeout: All probe commands have a 30-second timeout to prevent
//     runaway processes or resource exhaustion.
//
//  5. Trusted Image: The probe image (certsuite-probe) is maintained by the Red Hat
//     Best Practices team and contains only standard Linux utilities for inspection.
//     No custom binaries or scripts are included.
//
//  6. Explicit RBAC: Operators must grant pods/exec permissions explicitly. This
//     provides an audit trail of who can execute commands via the probe.
//
// # Checks Requiring Probe Access
//
// The following check categories require node-level access:
//   - platform: Node configuration, kernel parameters, OS details
//   - networking: iptables rules, routing tables, network interfaces
//   - performance: CPU governor, NUMA configuration, hugepages
//
// Checks that only inspect Kubernetes API objects (pods, services, etc.) do not
// require probe access and run directly in the operator.
package probe

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ProbeImage    = "quay.io/redhat-best-practices-for-k8s/certsuite-probe:v0.0.32"
	ProbeName     = "certsuite-probe"
	ProbeLabel    = "redhat-best-practices-for-k8s.com/app"
	ProbeLabelVal = "certsuite-probe"
	HostMountPath = "/host"
	HostMountName = "host-root"
)

// EnsureDaemonSet creates or updates the certsuite-probe DaemonSet in the given namespace.
func EnsureDaemonSet(ctx context.Context, c client.Client, namespace, image string) error {
	ds := desiredDaemonSet(namespace, image)

	var existing appsv1.DaemonSet
	err := c.Get(ctx, types.NamespacedName{Name: ProbeName, Namespace: namespace}, &existing)
	if errors.IsNotFound(err) {
		return c.Create(ctx, ds)
	}
	if err != nil {
		return err
	}

	// Only update if the spec has changed
	if !equality.Semantic.DeepEqual(existing.Spec, ds.Spec) {
		existing.Spec = ds.Spec
		return c.Update(ctx, &existing)
	}
	return nil // No update needed
}

// DeleteDaemonSet removes the certsuite-probe DaemonSet from the given namespace.
func DeleteDaemonSet(ctx context.Context, c client.Client, namespace string) error {
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ProbeName,
			Namespace: namespace,
		},
	}
	err := c.Delete(ctx, ds)
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}

// MapProbePods returns a map of node name to probe pod for all running probe pods.
func MapProbePods(ctx context.Context, c client.Client, namespace string) (map[string]*corev1.Pod, error) {
	var podList corev1.PodList
	if err := c.List(ctx, &podList,
		client.InNamespace(namespace),
		client.MatchingLabels{ProbeLabel: ProbeLabelVal},
	); err != nil {
		return nil, fmt.Errorf("listing probe pods: %w", err)
	}

	result := make(map[string]*corev1.Pod, len(podList.Items))
	for i := range podList.Items {
		pod := &podList.Items[i]
		if pod.Status.Phase == corev1.PodRunning {
			result[pod.Spec.NodeName] = pod
		}
	}
	return result, nil
}

// desiredDaemonSet constructs the certsuite-probe DaemonSet specification.
//
// The DaemonSet runs with elevated privileges to enable node-level compliance checks.
// See package documentation for security model and justification.
func desiredDaemonSet(namespace, image string) *appsv1.DaemonSet {
	privileged := true
	hostPathDir := corev1.HostPathDirectory
	labels := map[string]string{ProbeLabel: ProbeLabelVal}

	// Use provided image or fall back to default
	if image == "" {
		image = ProbeImage
	}

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ProbeName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					// HostNetwork: Required to inspect network interfaces, routing tables, and iptables
					HostNetwork: true,
					// HostIPC: Required to inspect shared memory and IPC namespaces
					HostIPC: true,
					// HostPID: Required to inspect host processes and validate runtime configurations
					HostPID: true,
					Tolerations: []corev1.Toleration{
						{
							Key:      "node-role.kubernetes.io/control-plane",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
						{
							Key:      "node-role.kubernetes.io/master",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
					},
					Containers: []corev1.Container{
						{
							Name:    ProbeName,
							Image:   image,
							Command: []string{"sleep", "infinity"},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      HostMountName,
									MountPath: HostMountPath,
									// ReadOnly: Mounted read-only to prevent any modifications to the host
									ReadOnly: true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: HostMountName,
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/",
									Type: &hostPathDir,
								},
							},
						},
					},
				},
			},
		},
	}
}
