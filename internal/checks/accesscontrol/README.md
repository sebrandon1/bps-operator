# Access Control Checks

Validates security contexts, Linux capabilities, host resource isolation, RBAC configuration, and namespace hygiene. These checks ensure workloads follow the principle of least privilege.

## Checks

| Check | Description | Remediation | Catalog Link |
|---|---|---|---|
| `access-control-sys-admin` | Verifies containers do not have SYS_ADMIN capability | Remove SYS_ADMIN from securityContext.capabilities.add | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-sys-admin-capability-check) |
| `access-control-net-admin` | Verifies containers do not have NET_ADMIN capability | Remove NET_ADMIN from securityContext.capabilities.add | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-net-admin-capability-check) |
| `access-control-net-raw` | Verifies containers do not have NET_RAW capability | Remove NET_RAW from securityContext.capabilities.add | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-net-raw-capability-check) |
| `access-control-ipc-lock` | Verifies containers do not have IPC_LOCK capability | Remove IPC_LOCK from securityContext.capabilities.add | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-ipc-lock-capability-check) |
| `access-control-bpf` | Verifies containers do not have BPF capability | Remove BPF from securityContext.capabilities.add | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-bpf-capability-check) |
| `access-control-host-network` | Verifies pods do not use HostNetwork | Set spec.hostNetwork to false | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-pod-host-network) |
| `access-control-host-path` | Verifies pods do not use HostPath volumes | Remove HostPath volumes from pod spec | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-pod-host-path) |
| `access-control-host-ipc` | Verifies pods do not use HostIPC | Set spec.hostIPC to false | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-pod-host-ipc) |
| `access-control-host-pid` | Verifies pods do not use HostPID | Set spec.hostPID to false | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-pod-host-pid) |
| `access-control-container-host-port` | Verifies containers do not use HostPort | Remove hostPort from container port definitions | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-container-host-port) |
| `access-control-non-root-user` | Verifies containers set runAsNonRoot to true | Set securityContext.runAsNonRoot to true | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-security-context-non-root-user-id-check) |
| `access-control-privilege-escalation` | Verifies containers set allowPrivilegeEscalation to false | Set securityContext.allowPrivilegeEscalation to false | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-security-context-privilege-escalation) |
| `access-control-read-only-filesystem` | Verifies containers set readOnlyRootFilesystem to true | Set securityContext.readOnlyRootFilesystem to true | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-security-context-read-only-file-system) |
| `access-control-1337-uid` | Verifies containers do not run as UID 1337 (reserved by Istio) | Use a UID other than 1337 | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-no-1337-uid) |
| `access-control-security-context` | Categorizes container security context (SCC classification) | Ensure containers do not require privileged SCC | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-security-context) |
| `access-control-service-account` | Verifies pods do not use the default service account | Create and assign a dedicated service account | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-pod-service-account) |
| `access-control-role-bindings` | Verifies RoleBindings do not reference non-target namespace ServiceAccounts | Ensure RoleBindings only reference ServiceAccounts from target namespaces | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-pod-role-bindings) |
| `access-control-cluster-role-bindings` | Verifies pod ServiceAccounts are not bound to ClusterRoleBindings | Use namespace-scoped RoleBindings instead of ClusterRoleBindings | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-cluster-role-bindings) |
| `access-control-automount-token` | Verifies pods do not automount service account tokens | Set automountServiceAccountToken to false on the pod or service account | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-pod-automount-service-account-token) |
| `access-control-nodeport-service` | Verifies services do not use NodePort type | Use ClusterIP or LoadBalancer service type instead | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-service-type) |
| `access-control-namespace` | Verifies pods run in allowed namespaces (not system namespaces) | Deploy workloads in dedicated namespaces, not default/kube-system/openshift-* | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-namespace) |
| `access-control-namespace-resource-quota` | Verifies namespace has a ResourceQuota defined | Create a ResourceQuota in the target namespace | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-namespace-resource-quota) |
| `access-control-one-process` | Verifies each container runs only one process | Ensure each container has a single main process | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-one-process-per-container) |
| `access-control-no-sshd` | Verifies no SSH daemons are running in containers | Remove SSH daemon from container images | [View](https://github.com/redhat-best-practices-for-k8s/certsuite/blob/main/CATALOG.md#access-control-ssh-daemons) |

## Check Types

- **Spec-based**: Most checks inspect Kubernetes object specs (pod specs, service specs, RBAC resources) and require only read access to the API server.
- **Probe-based**: `access-control-one-process` and `access-control-no-sshd` exec into containers via the probe DaemonSet to inspect running processes.

## Files

| File | Description |
|---|---|
| `register.go` | Registers all 24 access-control checks |
| `capabilities.go` | Capability checks (SYS_ADMIN, NET_ADMIN, NET_RAW, IPC_LOCK, BPF) |
| `host_checks.go` | Host resource checks (HostNetwork, HostPath, HostIPC, HostPID, HostPort) |
| `security_context.go` | Security context checks (non-root, privilege escalation, read-only FS, UID 1337, SCC classification) |
| `rbac_checks.go` | RBAC checks (service account, role bindings, cluster role bindings, automount token) |
| `services.go` | Service checks (NodePort) |
| `namespace.go` | Namespace checks (allowed namespaces, resource quotas) |
| `processes.go` | Probe-based process checks (one process, no SSHD) |
| `accesscontrol_test.go` | Unit tests |
