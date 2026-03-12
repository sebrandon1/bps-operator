# Lifecycle Checks

Validates container probes, lifecycle hooks, image pull policies, pod ownership, scheduling, high availability, and storage configuration. These checks ensure workloads are resilient and properly managed throughout their lifecycle.

## Checks

| Check | Description | Remediation | Catalog Link |
|---|---|---|---|
| `lifecycle-container-startup` | Verifies containers have a startupProbe defined | Add a startupProbe to the container spec | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-container-startup) |
| `lifecycle-container-readiness` | Verifies containers have a readinessProbe defined | Add a readinessProbe to the container spec | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-container-readiness) |
| `lifecycle-container-liveness` | Verifies containers have a livenessProbe defined | Add a livenessProbe to the container spec | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-container-liveness) |
| `lifecycle-container-prestop` | Verifies containers have a preStop lifecycle hook | Add a preStop lifecycle hook to the container spec | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-container-prestop) |
| `lifecycle-container-poststart` | Verifies containers have a postStart lifecycle hook | Add a postStart lifecycle hook to the container spec | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-container-poststart) |
| `lifecycle-image-pull-policy` | Verifies imagePullPolicy is Always or uses image digest | Set imagePullPolicy to Always or use an image reference with a digest | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-image-pull-policy) |
| `lifecycle-pod-owner-type` | Verifies pods are owned by ReplicaSet, StatefulSet, or DaemonSet | Deploy pods via Deployment, StatefulSet, or DaemonSet | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-pod-owner-type) |
| `lifecycle-pod-scheduling` | Verifies pods have scheduling directives (nodeSelector, affinity, or tolerations) | Add nodeSelector, affinity, or tolerations to the pod spec | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-pod-scheduling) |
| `lifecycle-pod-high-availability` | Verifies Deployments have replicas > 1 for high availability | Set spec.replicas to at least 2 | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-pod-high-availability) |
| `lifecycle-cpu-isolation` | Verifies CPU requests equal CPU limits (Guaranteed QoS for CPU) | Set CPU requests equal to CPU limits | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-cpu-isolation) |
| `lifecycle-affinity-required-pods` | Verifies pods have pod anti-affinity for high availability | Add podAntiAffinity to spread replicas across nodes | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-affinity-required-pods) |
| `lifecycle-pod-toleration-bypass` | Verifies pods do not tolerate NoExecute/NoSchedule master taints unnecessarily | Remove unnecessary tolerations for master/control-plane taints | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-pod-toleration-bypass) |
| `lifecycle-persistent-volume-reclaim-policy` | Verifies PersistentVolume reclaimPolicy is not Delete | Set persistentVolumeReclaimPolicy to Retain | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-persistent-volume-reclaim-policy) |
| `lifecycle-storage-provisioner` | Verifies StorageClass has a valid provisioner | Use a supported StorageClass provisioner | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#lifecycle-storage-provisioner) |

## Check Types

- **Spec-based**: All lifecycle checks inspect Kubernetes object specs (container specs, pod specs, PV/StorageClass specs) and require only read access to the API server.

## Files

| File | Description |
|---|---|
| `register.go` | Registers all 14 lifecycle checks |
| `probes.go` | Probe checks (startup, readiness, liveness) |
| `hooks.go` | Lifecycle hook checks (preStop, postStart) |
| `pods.go` | Pod spec checks (image pull policy, owner type, scheduling, HA, CPU isolation, affinity, tolerations, PV reclaim policy, storage provisioner) |
| `lifecycle_test.go` | Unit tests |
