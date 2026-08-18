# bps-operator

A Kubernetes operator that scans workloads against [certsuite](https://github.com/redhat-best-practices-for-k8s/certsuite) best practices and produces `BestPracticeResult` custom resources with per-check compliance status.

## Key Features

- **105 checks** across 9 categories from [redhat-best-practices-for-k8s/checks](https://github.com/redhat-best-practices-for-k8s/checks)
- **One-shot or periodic** scanning via `scanInterval` on the `BestPracticeScanner` CR
- **Label selectors** to target specific workloads
- **Remediation guidance** and [certsuite catalog](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md) links on every result
- **Kubernetes-native** -- results are CRs you can query with `kubectl`

## Quick Start

### Without Cloning

```bash
# Deploy the operator
kubectl apply -f https://raw.githubusercontent.com/sebrandon1/bps-operator/main/install.yaml

# Create a scanner in your namespace (scans all pods)
kubectl apply -f https://raw.githubusercontent.com/sebrandon1/bps-operator/main/config/samples/scanner_sample.yaml

# View results
kubectl get bestpracticeresults -n default
```

### With the Repo

```bash
git clone https://github.com/sebrandon1/bps-operator.git && cd bps-operator

# One-shot scan -- deploys operator in-cluster, runs all checks, shows results
make deploy-scan

# Or for continuous scanning every 5 minutes
make deploy-periodic-scan

# View results anytime
make show-results
make show-failures

# Clean up everything
make clean
```

## Guides

| Document | Description |
|---|---|
| [CRD API Reference](docs/crd-api.md) | `BestPracticeScanner` and `BestPracticeResult` field specs |
| [Configuration](docs/configuration.md) | Scanner CR examples and field reference |
| [Checks](docs/checks.md) | Check category summary (105 checks across 9 categories) |
| [Architecture](docs/architecture.md) | Project layout, check library, certsuite alignment |
| [Security Model](docs/security-model.md) | Probe DaemonSet privileges and security boundaries |
| [Development](docs/development.md) | Full `make` target reference |
| [Changelog](CHANGELOG.md) | Release history |

## Development

See [Development](docs/development.md) for build, test, lint, and deployment targets.
