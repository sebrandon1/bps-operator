# Platform Checks

Validates kernel boot parameters, hugepage configuration, sysctl settings, kernel taint status, service mesh usage, and cluster node count. These checks ensure the underlying platform has not been altered in unsupported ways.

## Checks

| Check | Description | Remediation | Catalog Link |
|---|---|---|---|
| `platform-boot-params` | Verifies no non-standard kernel boot parameters are set | Use MachineConfig to manage kernel boot parameters | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#platform-alteration-boot-params) |
| `platform-hugepages` | Verifies hugepage configuration is consistent | Configure hugepages via MachineConfig or performance profile | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#platform-alteration-hugepages-config) |
| `platform-sysctl` | Verifies sysctl settings are not modified outside MachineConfig | Use MachineConfig to manage sysctl settings | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#platform-alteration-sysctl-config) |
| `platform-tainted` | Verifies the kernel is not tainted | Investigate and resolve kernel taint causes | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#platform-alteration-tainted-node-kernel) |
| `platform-service-mesh-usage` | Verifies pods do not use service mesh (Istio) sidecars | Remove Istio sidecar injection if not required | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#platform-alteration-service-mesh-usage) |
| `platform-hugepages-2mi-only` | Verifies only 2Mi hugepages are used (not 1Gi) | Use 2Mi hugepages instead of 1Gi | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#platform-alteration-hugepages-2mi-only) |
| `platform-ocp-node-count` | Verifies cluster has minimum recommended number of worker nodes | Ensure cluster has at least 3 worker nodes | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#platform-alteration-ocp-node-count) |

## Check Types

- **Probe-based**: `platform-boot-params`, `platform-hugepages`, `platform-sysctl`, and `platform-tainted` exec into nodes via the probe DaemonSet to inspect kernel and system configuration.
- **Spec-based**: `platform-service-mesh-usage`, `platform-hugepages-2mi-only`, and `platform-ocp-node-count` inspect Kubernetes object specs and node lists without requiring exec access.

## Files

| File | Description |
|---|---|
| `register.go` | Registers all 7 platform checks |
| `boot_params.go` | Kernel boot parameter check |
| `hugepages.go` | Hugepage configuration checks (consistency and 2Mi-only) |
| `sysctl.go` | Sysctl settings check |
| `tainted.go` | Kernel taint check |
| `nodes.go` | Node-level checks (service mesh usage, node count) |
