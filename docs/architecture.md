# Architecture

## Project Layout

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

## Check Library

Check implementations are provided by the external [redhat-best-practices-for-k8s/checks](https://github.com/redhat-best-practices-for-k8s/checks) library.

## Certsuite Alignment

This operator implements a subset of the checks from the [certsuite](https://github.com/redhat-best-practices-for-k8s/certsuite) project as a Kubernetes-native operator. Each check's `CatalogID` maps directly to an entry in the certsuite [CATALOG.md](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md), enabling traceability between operator results and the upstream test catalog.

The operator is designed to run continuously in-cluster, providing real-time compliance feedback as workloads are deployed, rather than requiring a separate test execution step.
