# bps-operator

A Kubernetes operator that scans workloads against [certsuite](https://github.com/redhat-best-practices-for-k8s/certsuite) best practices and produces `BestPracticeResult` custom resources with per-check compliance status.

## Overview

bps-operator watches for `BestPracticeScanner` custom resources and runs a configurable set of best-practice checks against pods, services, and other resources in the target namespace. Each check produces a `BestPracticeResult` CR recording whether the resource is compliant, non-compliant, skipped, or errored, along with remediation guidance and a link to the corresponding [certsuite CATALOG.md](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md) entry.

## Quick Start

**Prerequisites**: A running Kubernetes or OpenShift cluster with `kubectl`/`oc` configured.

```bash
# One-shot scan — deploys operator in-cluster, runs all checks, shows results
make deploy-scan

# Or for continuous scanning every 5 minutes
make deploy-periodic-scan

# View results anytime
make show-results
make show-failures

# Clean up everything
make clean
```

## CRD API

### BestPracticeScanner

Defines a scan request.

| Field | Type | Description |
|---|---|---|
| `spec.targetNamespace` | `string` | Namespace to scan (defaults to the CR's namespace) |
| `spec.labelSelector` | `LabelSelector` | Filters which pods to scan |
| `spec.scanInterval` | `string` | Interval between scans (e.g. `5m`, `1h30m`); validated Go duration format; omit for one-shot |
| `spec.checks` | `[]string` | Specific checks to run (minimum 1 if specified); empty means all |
| `spec.suspend` | `bool` | Pauses scanning when `true` |

Status fields: `phase` (Idle/Scanning/Completed/Error), `lastScanTime`, `nextScanTime`, `summary` (total/compliant/nonCompliant/error/skipped counts).

### BestPracticeResult

Records the outcome of a single check.

| Field | Type | Description |
|---|---|---|
| `spec.scannerRef` | `string` | Name of the scanner that produced this result |
| `spec.checkName` | `string` | Unique check identifier |
| `spec.category` | `string` | Check category (e.g. `access-control`) |
| `spec.complianceStatus` | `string` | `Compliant`, `NonCompliant`, `Error`, or `Skipped` |
| `spec.reason` | `string` | Explanation of the result |
| `spec.remediation` | `string` | How to fix non-compliance |
| `spec.catalogURL` | `string` | Link to the certsuite catalog entry |
| `spec.details` | `[]ResourceDetail` | Per-resource compliance breakdown |

## Checks Summary

105 checks across 9 categories (provided by [redhat-best-practices-for-k8s/checks](https://github.com/redhat-best-practices-for-k8s/checks)):

| Category | Count |
|---|---|
| Access Control | 28 |
| Lifecycle | 19 |
| Platform Alteration | 15 |
| Operator | 12 |
| Networking | 11 |
| Performance | 9 |
| Observability | 5 |
| Affiliated Certification | 4 |
| Manageability | 2 |

## Usage

| Make Target | Description |
|---|---|
| `make build` | Build the operator binary |
| `make test` | Run unit tests with coverage |
| `make lint` | Run golangci-lint |
| `make install` | Install CRDs onto the cluster |
| `make deploy` | Deploy operator to the cluster (CRDs + RBAC + manager) |
| `make deploy-test` | Deploy test workloads only (no scanner) into `bps-test` namespace |
| `make deploy-scan` | Deploy operator, run one-shot scan, show results |
| `make deploy-periodic-scan` | Deploy operator, start periodic scan (5m interval) |
| `make scan` | Alias for `deploy-scan` |
| `make show-results` | Show scan results from the cluster |
| `make show-failures` | Show details for all non-compliant results |
| `make show-scan-yaml` | Print the one-shot scanner CR YAML |
| `make show-periodic-scan-yaml` | Print the periodic scanner CR YAML |
| `make clean` | Remove everything: test workloads, CRDs, namespace |
| `make build-image` | Build container image |
| `make manifests` | Regenerate CRD and RBAC manifests |
| `make generate` | Regenerate deepcopy functions |

## Configuration

Create a `BestPracticeScanner` CR to configure scanning:

```yaml
apiVersion: bps.redhat-best-practices-for-k8s.com/v1alpha1
kind: BestPracticeScanner
metadata:
  name: my-scanner
  namespace: my-app
spec:
  targetNamespace: my-app
  labelSelector:
    matchLabels:
      app: my-workload
  scanInterval: "10m"
  checks:
    - access-control-sys-admin
    - lifecycle-container-liveness
  suspend: false
```

- **targetNamespace**: Which namespace to scan. Defaults to the CR's own namespace.
- **labelSelector**: Filter pods by labels. Omit to scan all pods in the namespace.
- **scanInterval**: How often to re-scan (e.g., "5m", "1h30m", "10s"). Must be a valid Go duration string. Omit for a one-shot scan. The API validates this format at admission time, rejecting invalid durations like "5mins" or "1 hour".
- **checks**: Run only specific checks by name. Omit to run all checks.
- **suspend**: Set to `true` to pause periodic scanning.

## Architecture

```
cmd/                     Main entrypoint
api/v1alpha1/            CRD type definitions (BestPracticeScanner, BestPracticeResult)
internal/
  controller/            Reconciler for BestPracticeScanner CRs
  scanner/               Orchestrates check execution and result creation
  certification/         Red Hat container certification validation
  probe/                 Probe DaemonSet for exec-based checks
config/
  crd/bases/             Generated CRD manifests
  rbac/                  RBAC manifests
  manager/               Operator Deployment manifest
  samples/               Example CRs and test workloads
```

Check implementations are provided by the external [redhat-best-practices-for-k8s/checks](https://github.com/redhat-best-practices-for-k8s/checks) library.

## Security Model

### Probe DaemonSet Privileges

Some compliance checks require node-level access to verify host configurations (kernel parameters, network settings, etc.). For these checks, the operator deploys a privileged DaemonSet called `certsuite-probe`.

**Why Privileged Access is Required:**

The probe runs with elevated privileges to enable checks such as:
- Kernel parameter verification (sysctl settings)
- Host file inspection (/etc, /proc, /sys)
- Network configuration validation (iptables, routing tables)
- Container runtime inspection
- Security context validation

**Security Boundaries:**

1. **Namespace Isolation**: Probe pods run only in the operator namespace (`bps-operator-system`), not in user workload namespaces being scanned.

2. **Read-Only Host Access**: The host root filesystem is mounted read-only at `/host`, preventing any modifications to nodes.

3. **No Automated Execution**: The probe container runs `sleep infinity` with no automated code execution. Commands are executed only via explicit Kubernetes RBAC-controlled `pods/exec` API calls.

4. **Execution Timeout**: All probe commands have a 30-second timeout to prevent runaway processes.

5. **Trusted Image**: The probe image ([certsuite-probe](https://quay.io/repository/redhat-best-practices-for-k8s/certsuite-probe)) is maintained by the Red Hat Best Practices team and contains only standard Linux utilities.

6. **RBAC Audit Trail**: Operators must grant `pods/exec` permissions explicitly via ClusterRole. All command executions are logged by the Kubernetes API server for audit purposes.

**Checks Requiring Probe Access:**

- **Platform checks**: Node configuration, OS details, kernel parameters
- **Networking checks**: iptables rules, routing tables, interface configuration
- **Performance checks**: CPU governor, NUMA topology, hugepages settings

Checks that only inspect Kubernetes API objects (pods, services, RBAC, etc.) run directly in the operator without elevated privileges.

For detailed security documentation, see [internal/probe/daemonset.go](internal/probe/daemonset.go).

## Certsuite Alignment

This operator implements a subset of the checks from the [certsuite](https://github.com/redhat-best-practices-for-k8s/certsuite) project as a Kubernetes-native operator. Each check's `CatalogID` maps directly to an entry in the certsuite [CATALOG.md](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md), enabling traceability between operator results and the upstream test catalog.

The operator is designed to run continuously in-cluster, providing real-time compliance feedback as workloads are deployed, rather than requiring a separate test execution step.
