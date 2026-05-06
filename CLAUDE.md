# CLAUDE.md

## Project Overview

Kubernetes operator that continuously scans workload compliance against Red Hat best practices using the certsuite framework. Watches BestPracticeScanner custom resources and produces BestPracticeResult CRs with compliance status, remediation guidance, and catalog links for 105 checks across 9 categories.

## Build & Run

```bash
make build            # Build operator binary to bin/manager
make manifests        # Regenerate CRD and RBAC manifests
make generate         # Regenerate deepcopy functions
make build-image      # Build Docker container image
```

## Testing

```bash
make test             # Unit tests with coverage (cover.out)
make test-e2e         # E2E tests against Kind cluster
make coverage-html    # HTML coverage report
```

## Linting

```bash
make lint             # golangci-lint (govet, errcheck, staticcheck, unused, ineffassign)
make fmt              # go fmt
make vet              # go vet
```

## Key Architecture

- **CRDs:** BestPracticeScanner (scan config), BestPracticeResult (check outcome)
- **Namespace:** bps-operator-system
- Deploys privileged DaemonSet (certsuite-probe) for host-level checks
- Auto-detects OpenShift via oc command
