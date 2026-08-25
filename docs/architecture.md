# Architecture

## Project Layout

```
cmd/                     Main entrypoint
internal/
  controller/            Reconciler for BestPracticeScanner CRs
  scanner/               Orchestrates check execution and result creation
  certification/         Red Hat container certification validation
  probe/                 Probe DaemonSet for exec-based checks
  metrics/               Prometheus metrics
config/
  crd/bases/             Generated CRD manifests
  rbac/                  RBAC manifests
  manager/               Operator Deployment manifest
  samples/               Example CRs and test workloads
  grafana/               Grafana dashboard JSON for Prometheus metrics
  webhook/               ValidatingWebhookConfiguration manifest
```

CRD type definitions (`BestPracticeScanner`, `BestPracticeResult`) live in the external [checks-types](https://github.com/redhat-best-practices-for-k8s/checks-types) library. Both CRs use API group `bps.openshift.io/v1alpha1`.

## Prometheus Metrics

The operator exposes four Prometheus metrics via the controller-runtime metrics endpoint (default `:8080/metrics`):

| Metric | Type | Labels | Description |
|---|---|---|---|
| `bps_scan_duration_seconds` | Histogram | `scanner`, `namespace` | Duration of each full scan |
| `bps_scans_total` | Counter | `scanner`, `namespace` | Total completed scans |
| `bps_check_results` | Gauge | `scanner`, `namespace`, `status` | Check result counts by status from the most recent scan |
| `bps_check_duration_seconds` | Histogram | `check`, `category` | Duration of individual checks |

A pre-built Grafana dashboard is available at `config/grafana/dashboard.json`. Import it into any Grafana instance connected to Prometheus.

## Check Library

Check implementations are provided by the external [redhat-best-practices-for-k8s/checks](https://github.com/redhat-best-practices-for-k8s/checks) library.

## Certsuite Alignment

This operator implements a subset of the checks from the [certsuite](https://github.com/redhat-best-practices-for-k8s/certsuite) project as a Kubernetes-native operator. Each check's `CatalogID` maps directly to an entry in the certsuite [CATALOG.md](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md), enabling traceability between operator results and the upstream test catalog.

The operator is designed to run continuously in-cluster, providing real-time compliance feedback as workloads are deployed, rather than requiring a separate test execution step.
