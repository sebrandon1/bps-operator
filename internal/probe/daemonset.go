package probe

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ProbeImage     = "quay.io/redhat-best-practices-for-k8s/certsuite-probe:v0.0.32"
	ProbeName      = "certsuite-probe"
	ProbeLabel     = "redhat-best-practices-for-k8s.com/app"
	ProbeLabelVal  = "certsuite-probe"
	HostMountPath  = "/host"
	HostMountName  = "host-root"
)

// EnsureDaemonSet creates or updates the certsuite-probe DaemonSet in the given namespace.
func EnsureDaemonSet(ctx context.Context, c client.Client, namespace string) error {
	ds := desiredDaemonSet(namespace)

	var existing appsv1.DaemonSet
	err := c.Get(ctx, types.NamespacedName{Name: ProbeName, Namespace: namespace}, &existing)
	if errors.IsNotFound(err) {
		return c.Create(ctx, ds)
	}
	if err != nil {
		return err
	}

	// Update the spec if it already exists
	existing.Spec = ds.Spec
	return c.Update(ctx, &existing)
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

func desiredDaemonSet(namespace string) *appsv1.DaemonSet {
	privileged := true
	hostPathDir := corev1.HostPathDirectory
	labels := map[string]string{ProbeLabel: ProbeLabelVal}

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
					HostNetwork: true,
					HostIPC:     true,
					HostPID:     true,
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
							Image:   ProbeImage,
							Command: []string{"sleep", "infinity"},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      HostMountName,
									MountPath: HostMountPath,
									ReadOnly:  true,
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
