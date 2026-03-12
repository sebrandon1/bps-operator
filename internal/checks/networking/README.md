# Networking Checks

Validates dual-stack service configuration, network policy enforcement, and port usage. These checks ensure workloads follow networking best practices for Kubernetes and OpenShift environments.

## Checks

| Check | Description | Remediation | Catalog Link |
|---|---|---|---|
| `networking-dual-stack-service` | Verifies services support dual-stack (both IPv4 and IPv6) | Set spec.ipFamilyPolicy to PreferDualStack or RequireDualStack | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#networking-dual-stack-service) |
| `networking-network-policy-deny-all` | Verifies a default-deny NetworkPolicy exists in the namespace | Create a NetworkPolicy that denies all ingress and egress by default | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#networking-network-policy-deny-all) |
| `networking-reserved-partner-ports` | Verifies containers do not bind to reserved partner ports (22222, 22623, 22624) | Use non-reserved ports for container services | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#networking-reserved-partner-ports) |
| `networking-ocp-reserved-ports-usage` | Verifies containers do not use OpenShift reserved ports (22623, 22624) | Avoid using OpenShift reserved ports | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#networking-ocp-reserved-ports-usage) |

## Check Types

- **Spec-based**: All networking checks inspect Kubernetes object specs (service specs, network policies, container port definitions) and require only read access to the API server.

## Files

| File | Description |
|---|---|
| `register.go` | Registers all 4 networking checks |
| `services.go` | Service checks (dual-stack, reserved ports, OCP reserved ports) |
| `policies.go` | Network policy checks (default-deny) |
| `networking_test.go` | Unit tests |
