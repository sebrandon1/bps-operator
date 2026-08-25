# Development

Run `make help` for the full annotated target list.

## Build, Test, and Lint

| Make Target | Description |
|---|---|
| `make build` | Build the operator binary |
| `make test` | Run unit tests with coverage |
| `make test-e2e` | Run e2e tests against a Kind cluster |
| `make coverage-html` | Generate HTML coverage report from `cover.out` |
| `make lint` | Run golangci-lint |
| `make fmt` | Run `go fmt` |
| `make vet` | Run `go vet` |
| `make list-checks` | Print all valid check names |
| `make generate` | Regenerate deepcopy functions |
| `make manifests` | Regenerate CRD and RBAC manifests |
| `make build-image` | Build container image |
| `make push-image` | Push container image to registry |

## Installation and Deployment

| Make Target | Description |
|---|---|
| `make install-yaml` | Generate `install.yaml` with all resources for remote deployment |
| `make install` | Install CRDs onto the cluster |
| `make uninstall` | Remove CRDs from the cluster |
| `make deploy` | Deploy operator to the cluster (CRDs + RBAC + manager) |
| `make undeploy` | Remove the operator from the cluster |

## Scanning

| Make Target | Description |
|---|---|
| `make deploy-test` | Deploy test workloads only (no scanner) into `bps-test` namespace |
| `make deploy-scan` | Deploy operator, run one-shot scan, show results |
| `make deploy-periodic-scan` | Deploy operator, start periodic scan (5m interval) |
| `make scan` | Alias for `deploy-scan` |
| `make show-results` | Show scan results from the cluster |
| `make show-failures` | Show details for all non-compliant results |
| `make show-scan-yaml` | Print the one-shot scanner CR YAML |
| `make show-periodic-scan-yaml` | Print the periodic scanner CR YAML |
| `make undeploy-test` | Remove test workloads, scanner, and `bps-test` namespace |
| `make clean` | Remove everything: test workloads, CRDs, namespace |

## Release

| Make Target | Description |
|---|---|
| `make release-tag` | Create and push a release tag (usage: `make release-tag VERSION=v0.0.4`) |
