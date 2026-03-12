# Observability Checks

Validates CRD status subresources, termination message policies, and pod disruption budgets. These checks ensure workloads provide adequate observability and disruption management.

## Checks

| Check | Description | Remediation | Catalog Link |
|---|---|---|---|
| `observability-crd-status` | Verifies CRDs define a .status subresource | Add status subresource to the CRD spec versions | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#observability-crd-status) |
| `observability-termination-policy` | Verifies containers set terminationMessagePolicy to FallbackToLogsOnError | Set terminationMessagePolicy to FallbackToLogsOnError | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#observability-termination-policy) |
| `observability-pod-disruption-budget` | Verifies PodDisruptionBudgets exist for HA workloads | Create a PodDisruptionBudget for deployments with replicas > 1 | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#observability-pod-disruption-budget) |

## Check Types

- **Spec-based**: All observability checks inspect Kubernetes object specs (CRD definitions, container specs, PDB resources) and require only read access to the API server.

## Files

| File | Description |
|---|---|
| `register.go` | Registers all 3 observability checks |
| `crd_status.go` | CRD status subresource check |
| `termination_policy.go` | Termination message policy check |
| `pdb.go` | PodDisruptionBudget check |
| `observability_test.go` | Unit tests |
