# Manageability Checks

Validates container port naming conventions and image tag usage. These checks ensure workloads follow naming standards and use reproducible image references.

## Checks

| Check | Description | Remediation | Catalog Link |
|---|---|---|---|
| `manageability-container-port-name-format` | Verifies container port names follow IANA naming conventions | Use IANA-compliant port names (lowercase, alphanumeric, hyphens, max 15 chars) | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#manageability-container-port-name-format) |
| `manageability-containers-image-tag` | Verifies container images use a digest or specific tag (not :latest) | Use a specific image tag or digest reference instead of :latest | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#manageability-containers-image-tag) |

## Check Types

- **Spec-based**: Both checks inspect container specs (port definitions and image references) and require only read access to the API server.

## Files

| File | Description |
|---|---|
| `register.go` | Registers both manageability checks |
| `containers.go` | Container checks (port name format, image tag) |
| `manageability_test.go` | Unit tests |
