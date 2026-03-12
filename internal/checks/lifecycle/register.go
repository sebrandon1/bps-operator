package lifecycle

import "github.com/sebrandon1/bps-operator/internal/checks"

func init() {
	// Probe checks
	checks.Register(checks.CheckInfo{
		Name: "lifecycle-container-startup", Category: "lifecycle",
		Description: "Verifies containers have a startupProbe defined",
		Remediation: "Add a startupProbe to the container spec",
		CatalogID:   "lifecycle-container-startup",
		Fn:          CheckStartupProbe,
	})
	checks.Register(checks.CheckInfo{
		Name: "lifecycle-container-readiness", Category: "lifecycle",
		Description: "Verifies containers have a readinessProbe defined",
		Remediation: "Add a readinessProbe to the container spec",
		CatalogID:   "lifecycle-container-readiness",
		Fn:          CheckReadinessProbe,
	})
	checks.Register(checks.CheckInfo{
		Name: "lifecycle-container-liveness", Category: "lifecycle",
		Description: "Verifies containers have a livenessProbe defined",
		Remediation: "Add a livenessProbe to the container spec",
		CatalogID:   "lifecycle-container-liveness",
		Fn:          CheckLivenessProbe,
	})

	// Lifecycle hook checks
	checks.Register(checks.CheckInfo{
		Name: "lifecycle-container-prestop", Category: "lifecycle",
		Description: "Verifies containers have a preStop lifecycle hook",
		Remediation: "Add a preStop lifecycle hook to the container spec",
		CatalogID:   "lifecycle-container-prestop",
		Fn:          CheckPreStop,
	})
	checks.Register(checks.CheckInfo{
		Name: "lifecycle-container-poststart", Category: "lifecycle",
		Description: "Verifies containers have a postStart lifecycle hook",
		Remediation: "Add a postStart lifecycle hook to the container spec",
		CatalogID:   "lifecycle-container-poststart",
		Fn:          CheckPostStart,
	})

	// Pod spec checks
	checks.Register(checks.CheckInfo{
		Name: "lifecycle-image-pull-policy", Category: "lifecycle",
		Description: "Verifies imagePullPolicy is Always or uses image digest",
		Remediation: "Set imagePullPolicy to Always or use an image reference with a digest",
		CatalogID:   "lifecycle-image-pull-policy",
		Fn:          CheckImagePullPolicy,
	})
	checks.Register(checks.CheckInfo{
		Name: "lifecycle-pod-owner-type", Category: "lifecycle",
		Description: "Verifies pods are owned by ReplicaSet, StatefulSet, or DaemonSet",
		Remediation: "Deploy pods via Deployment, StatefulSet, or DaemonSet",
		CatalogID:   "lifecycle-pod-owner-type",
		Fn:          CheckPodOwnerType,
	})
	checks.Register(checks.CheckInfo{
		Name: "lifecycle-pod-scheduling", Category: "lifecycle",
		Description: "Verifies pods have scheduling directives (nodeSelector, affinity, or tolerations)",
		Remediation: "Add nodeSelector, affinity, or tolerations to the pod spec",
		CatalogID:   "lifecycle-pod-scheduling",
		Fn:          CheckPodScheduling,
	})
	checks.Register(checks.CheckInfo{
		Name: "lifecycle-pod-high-availability", Category: "lifecycle",
		Description: "Verifies Deployments have replicas > 1 for high availability",
		Remediation: "Set spec.replicas to at least 2",
		CatalogID:   "lifecycle-pod-high-availability",
		Fn:          CheckHighAvailability,
	})
	checks.Register(checks.CheckInfo{
		Name: "lifecycle-cpu-isolation", Category: "lifecycle",
		Description: "Verifies CPU requests equal CPU limits (Guaranteed QoS for CPU)",
		Remediation: "Set CPU requests equal to CPU limits",
		CatalogID:   "lifecycle-cpu-isolation",
		Fn:          CheckCPUIsolation,
	})
	checks.Register(checks.CheckInfo{
		Name: "lifecycle-affinity-required-pods", Category: "lifecycle",
		Description: "Verifies pods have pod anti-affinity for high availability",
		Remediation: "Add podAntiAffinity to spread replicas across nodes",
		CatalogID:   "lifecycle-affinity-required-pods",
		Fn:          CheckAffinityRequired,
	})
	checks.Register(checks.CheckInfo{
		Name: "lifecycle-pod-toleration-bypass", Category: "lifecycle",
		Description: "Verifies pods do not tolerate NoExecute/NoSchedule master taints unnecessarily",
		Remediation: "Remove unnecessary tolerations for master/control-plane taints",
		CatalogID:   "lifecycle-pod-toleration-bypass",
		Fn:          CheckTolerationBypass,
	})
	checks.Register(checks.CheckInfo{
		Name: "lifecycle-persistent-volume-reclaim-policy", Category: "lifecycle",
		Description: "Verifies PersistentVolume reclaimPolicy is not Delete",
		Remediation: "Set persistentVolumeReclaimPolicy to Retain",
		CatalogID:   "lifecycle-persistent-volume-reclaim-policy",
		Fn:          CheckPVReclaimPolicy,
	})
	checks.Register(checks.CheckInfo{
		Name: "lifecycle-storage-provisioner", Category: "lifecycle",
		Description: "Verifies StorageClass has a valid provisioner",
		Remediation: "Use a supported StorageClass provisioner",
		CatalogID:   "lifecycle-storage-provisioner",
		Fn:          CheckStorageProvisioner,
	})
}
