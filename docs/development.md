# Development

## Make Targets

| Make Target | Description |
|---|---|
| `make build` | Build the operator binary |
| `make test` | Run unit tests with coverage |
| `make lint` | Run golangci-lint |
| `make install` | Install CRDs onto the cluster |
| `make install-yaml` | Generate `install.yaml` with all resources for remote deployment |
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
