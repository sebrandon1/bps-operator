# Changelog

All notable changes to bps-operator will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Webhook validation for `BestPracticeScanner` CRD: rejects invalid `scanInterval` formats and unknown check names at admission time (#164)
- Prometheus metric `bps_check_duration_seconds` for per-check execution timing with `check` and `category` labels (#162)
- Grafana dashboard JSON for Prometheus metrics at `config/grafana/dashboard.json` (#163)
- `make list-checks` target to print all valid check names (#160)

### Changed
- Retry with backoff for Pyxis API failures (#161)
- `make show-failures` now includes remediation guidance in output (#159)
- Fixed stale documentation across README and `docs/` (apiVersion, check names, project layout, CRD fields, make targets)
- Updated CHANGELOG through v0.0.17

## [0.0.17] - 2026-07-13

### Added
- OCP 4.20 E2E test workflow (#113)
- Parallel lint CI jobs (#115)

### Changed
- Simplified README by extracting content to `docs/` (#110)
- Bumped checks library to v0.0.26 (#108)
- Bumped checks-types to v1.0.1 (#107)
- Upgraded Go toolchain to 1.26.5 to resolve govulncheck CVE GO-2026-5856 (#119)
- Bumped k8s, OLM, Docker, and GitHub Actions dependencies (#109, #111, #112, #114, #116, #117, #118)

## [0.0.16] - 2026-06-08

### Changed
- Upgraded Go to 1.26.3 with toolchain directive (#101)
- Updated Go toolchain to 1.26.4 (#104)
- Bumped Docker and GitHub Actions dependencies (#102, #103)

## [0.0.15] - 2026-05-29

### Changed
- Bumped OLM to v0.43.0 (#99)
- Bumped k8s dependencies to v0.36.1 (#96)
- Bumped Docker and GitHub Actions dependencies (#95, #97, #98, #100)

## [0.0.14] - 2026-05-08

### Added
- `CLAUDE.md` for AI-assisted development (#92)

### Changed
- Upgraded Go to 1.26.3 (#93)
- Bumped k8s and GitHub Actions dependencies (#90, #94)

## [0.0.13] - 2026-05-02

### Changed
- Bumped checks library to v0.0.25 (#91)
- Bumped k8s dependencies to v0.36.0 (#90)
- Bumped GitHub Actions dependencies (#89)

## [0.0.12] - 2026-04-24

### Changed
- Bumped GitHub Actions dependencies (#87)

## [0.0.11] - 2026-04-20

### Added
- `--result-ttl` flag for automatic cleanup of expired scan results (#86)
- Automated release workflow triggered by version tags (#85)
- `VERSION` variable and `make release-tag` target in Makefile (#85)

### Changed
- Increased test coverage from 49% to 65% with 13 new controller and probe tests (#84)

## [0.0.10] - 2026-04-20

### Added
- `install.yaml` for no-clone deployment (#76)
- Networking-only and label selector scanner sample CRs (#78)

### Changed
- Migrated CRD types to shared `checks-types` library (#83)
- Deploy operator in-cluster and improve quick start workflow (#75)
- Use dot separator for result names and add godoc (#74)
- Use prebuilt checks-qe container image for e2e tests (#77)
- Bumped checks library to v0.0.24 (#82)
- Bumped k8s dependencies (#81), actions (#80), spdystream (#79)

## [0.0.9] - 2026-04-14

### Added
- Checks-qe e2e workflows for checks library validation (#73)

### Changed
- Replaced blocking probe wait with non-blocking requeue (#72)
- Bumped checks library to v0.0.23 (#71)
- Bumped OLM dependency (#70)

## [0.0.8] - 2026-04-08

### Changed
- Upgraded Go to 1.26.2 (#67)
- Bumped checks library to v0.0.7 (#68)
- Bumped operator-framework/api to v0.42.0 (#64)
- Bumped GitHub Actions (#66)

## [0.0.7] - 2026-03-24

### Changed
- Made catalog URL base configurable (#62)
- Replaced blank imports with explicit check registration (#63)
- Simplified discovery and metrics code (#61)

## [0.0.6] - 2026-03-24

### Changed
- Bumped OLM and k8s dependencies (#56, #57)
- Restricted e2e-ocp workflow to upstream repo (#59, #60)
- Bumped distroless base image (#54)

## [0.0.5] - 2026-03-24

### Added
- Kubernetes Events and Status Conditions (#53)
- Reduced code duplication in resource discovery (#52)

## [0.0.4] - 2026-03-23

### Added
- Prometheus metrics: `bps_scan_duration_seconds`, `bps_scans_total`, `bps_check_results` (#47)
- OCP 4.21 E2E test workflow (#44)
- Unit tests for flag parsing and scheme registration (#49)
- `--node-name` flag as alternative to `NODE_NAME` env var (#49)

### Changed
- Upgraded checks library from v0.0.3 to v0.0.4 (#50)
- Refactored `Reconcile` into smaller methods (#46)
- Refactored `main()` into testable functions (#49)
- Wired `K8sClientset` and `ScaleClient` (#42)
- Removed resource limits from operator deployment (#43)
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

[Unreleased]: https://github.com/sebrandon1/bps-operator/compare/v0.0.17...HEAD
[0.0.17]: https://github.com/sebrandon1/bps-operator/compare/v0.0.16...v0.0.17
[0.0.16]: https://github.com/sebrandon1/bps-operator/compare/v0.0.15...v0.0.16
[0.0.15]: https://github.com/sebrandon1/bps-operator/compare/v0.0.14...v0.0.15
[0.0.14]: https://github.com/sebrandon1/bps-operator/compare/v0.0.13...v0.0.14
[0.0.13]: https://github.com/sebrandon1/bps-operator/compare/v0.0.12...v0.0.13
[0.0.12]: https://github.com/sebrandon1/bps-operator/compare/v0.0.11...v0.0.12
[0.0.11]: https://github.com/sebrandon1/bps-operator/compare/v0.0.10...v0.0.11
[0.0.10]: https://github.com/sebrandon1/bps-operator/compare/v0.0.9...v0.0.10
[0.0.9]: https://github.com/sebrandon1/bps-operator/compare/v0.0.8...v0.0.9
[0.0.8]: https://github.com/sebrandon1/bps-operator/compare/v0.0.7...v0.0.8
[0.0.7]: https://github.com/sebrandon1/bps-operator/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/sebrandon1/bps-operator/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/sebrandon1/bps-operator/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/sebrandon1/bps-operator/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/sebrandon1/bps-operator/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/sebrandon1/bps-operator/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/sebrandon1/bps-operator/releases/tag/v0.0.1
