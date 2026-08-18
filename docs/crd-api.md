# CRD API Reference

API group: **`bps.openshift.io/v1alpha1`**

## BestPracticeScanner

Defines a scan request.

| Field | Type | Description |
|---|---|---|
| `spec.targetNamespace` | `string` | Namespace to scan (defaults to the CR's namespace) |
| `spec.labelSelector` | `LabelSelector` | Filters which pods to scan |
| `spec.scanInterval` | `string` | Interval between scans (e.g. `5m`, `1h30m`); validated Go duration format; omit for one-shot |
| `spec.checks` | `[]string` | Specific checks to run (minimum 1 if specified); empty means all |
| `spec.suspend` | `bool` | Pauses scanning when `true` |

Status fields: `phase` (Idle/Scanning/Completed/Error), `lastScanTime`, `nextScanTime`, `summary` (total/compliant/nonCompliant/error/skipped counts), `conditions` (Kubernetes-style status conditions such as `ScanComplete`).

## BestPracticeResult

Records the outcome of a single check.

| Field | Type | Description |
|---|---|---|
| `spec.scannerRef` | `string` | Name of the scanner that produced this result |
| `spec.checkName` | `string` | Unique check identifier |
| `spec.category` | `string` | Check category (e.g. `access-control`) |
| `spec.description` | `string` | What the check verifies |
| `spec.complianceStatus` | `string` | `Compliant`, `NonCompliant`, `Error`, or `Skipped` |
| `spec.reason` | `string` | Explanation of the result |
| `spec.remediation` | `string` | How to fix non-compliance |
| `spec.catalogURL` | `string` | Link to the certsuite catalog entry |
| `spec.details` | `[]ResourceDetail` | Per-resource compliance breakdown |
