# Security Model

## Probe DaemonSet Privileges

Some compliance checks require node-level access to verify host configurations (kernel parameters, network settings, etc.). For these checks, the operator deploys a privileged DaemonSet called `certsuite-probe`.

### Why Privileged Access is Required

The probe runs with elevated privileges to enable checks such as:
- Kernel parameter verification (sysctl settings)
- Host file inspection (/etc, /proc, /sys)
- Network configuration validation (iptables, routing tables)
- Container runtime inspection
- Security context validation

### Security Boundaries

1. **Namespace Isolation**: Probe pods run only in the operator namespace (`bps-operator-system`), not in user workload namespaces being scanned.

2. **Read-Only Host Access**: The host root filesystem is mounted read-only at `/host`, preventing any modifications to nodes.

3. **No Automated Execution**: The probe container runs `sleep infinity` with no automated code execution. Commands are executed only via explicit Kubernetes RBAC-controlled `pods/exec` API calls.

4. **Execution Timeout**: All probe commands have a 30-second timeout to prevent runaway processes.

5. **Trusted Image**: The probe image ([certsuite-probe](https://quay.io/repository/redhat-best-practices-for-k8s/certsuite-probe)) is maintained by the Red Hat Best Practices team and contains only standard Linux utilities.

6. **RBAC Audit Trail**: Operators must grant `pods/exec` permissions explicitly via ClusterRole. All command executions are logged by the Kubernetes API server for audit purposes.

### Checks Requiring Probe Access

- **Platform checks**: Node configuration, OS details, kernel parameters
- **Networking checks**: iptables rules, routing tables, interface configuration
- **Performance checks**: CPU governor, NUMA topology, hugepages settings

Checks that only inspect Kubernetes API objects (pods, services, RBAC, etc.) run directly in the operator without elevated privileges.

For detailed security documentation, see [internal/probe/daemonset.go](../internal/probe/daemonset.go).
