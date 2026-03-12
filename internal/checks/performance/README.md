# Performance Checks

Validates CPU allocation, probe types for real-time workloads, and memory limits. These checks ensure workloads are configured for optimal performance, particularly in telco and latency-sensitive environments.

## Checks

| Check | Description | Remediation | Catalog Link |
|---|---|---|---|
| `performance-exclusive-cpu-pool` | Verifies containers requesting whole CPUs use exclusive CPU pool (Guaranteed QoS) | Set CPU requests equal to limits with whole-number values for exclusive CPU allocation | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#performance-exclusive-cpu-pool) |
| `performance-rt-apps-no-exec-probes` | Verifies real-time containers do not use exec probes | Use httpGet or tcpSocket probes instead of exec for RT workloads | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#performance-rt-apps-no-exec-probes) |
| `performance-limit-memory-allocation` | Verifies containers have memory limits set | Set resources.limits.memory on all containers | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#performance-limit-memory-allocation) |

## Check Types

- **Spec-based**: All performance checks inspect container resource specifications and probe configurations. They require only read access to the API server.

## Files

| File | Description |
|---|---|
| `register.go` | Registers all 3 performance checks |
| `resources.go` | Resource checks (exclusive CPU pool, exec probes for RT apps, memory limits) |
| `performance_test.go` | Unit tests |
