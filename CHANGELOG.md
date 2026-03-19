# Changelog

All notable changes to bps-operator will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Prometheus metrics: `bps_scan_duration_seconds`, `bps_scans_total`, `bps_check_results` (#47)
- OCP 4.21 E2E test workflow using quick-ocp/CRC with OpenShift internal registry (#44)
- Unit tests for `cmd/main.go` flag parsing and scheme registration (#49)
- `--node-name` flag as alternative to `NODE_NAME` env var (#49)

### Changed
- Upgraded checks library from v0.0.3 to v0.0.4 (#50)
- Refactored `Reconcile` into smaller methods: `enforceUniqueness`, `discoverResources`, `runChecks`, `upsertResult`, `completeScan` (#46)
- Refactored `main()` into testable `parseFlags()` and `run()` functions (#49)
- Wired `K8sClientset` and `ScaleClient` to fix 5 lifecycle/observability check errors (#42)
- Removed resource limits from operator deployment to prevent OOMKills (#43)
- Bumped google.golang.org/grpc from 1.78.0 to 1.79.3 (#48)

## [0.0.3] - 2026-03-18

### Changed
- Upgraded checks library from v0.0.2 to v0.0.3 (57 -> 105 checks)
- Added new `affiliated-certification` category with 4 checks (container, operator, helm chart certification via Red Hat Pyxis API)
- Added discovery for 15+ new resource types: Roles, Helm chart releases, CSVs, CatalogSources, Subscriptions, PackageManifests, ClusterVersion, ClusterOperators, APIRequestCounts, NetworkAttachmentDefinitions, SR-IOV resources, and scalable CRD instances
- Added graceful handling for OpenShift/OLM-specific APIs (skipped on vanilla Kubernetes)
- Added K8s version and OpenShift version detection
- Added NODE_NAME downward API env var for scanner pod node awareness
- Added `--certification-api-url` flag for configurable Pyxis API endpoint
- Registered OpenShift, OLM, and network-attachment-definition schemes

## [0.0.2] - 2026-03-16

### Changed
- Updated container image registry from quay.io/redhat-best-practices-for-k8s to quay.io/bapalm
- Added disk-pressure toleration to operator deployment to support constrained development environments like CRC

## [0.0.1] - 2026-03-13

Initial release.

### Added
- Kubernetes operator for running best-practice compliance checks
- 57 checks across 7 categories via checks v0.0.1 library
- BestPracticeScanner CRD for triggering scans (one-shot and periodic)
- BestPracticeResult CRD for storing compliance results
- Multi-arch container image (amd64/arm64)
- E2E tests in CI

[Unreleased]: https://github.com/sebrandon1/bps-operator/compare/v0.0.3...HEAD
[0.0.3]: https://github.com/sebrandon1/bps-operator/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/sebrandon1/bps-operator/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/sebrandon1/bps-operator/releases/tag/v0.0.1
