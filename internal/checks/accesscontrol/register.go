package accesscontrol

import "github.com/sebrandon1/bps-operator/internal/checks"

func init() {
	// Capability checks
	checks.Register(checks.CheckInfo{
		Name: "access-control-sys-admin", Category: "access-control",
		Description: "Verifies containers do not have SYS_ADMIN capability",
		Remediation: "Remove SYS_ADMIN from securityContext.capabilities.add",
		Fn:          CheckSysAdmin,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-net-admin", Category: "access-control",
		Description: "Verifies containers do not have NET_ADMIN capability",
		Remediation: "Remove NET_ADMIN from securityContext.capabilities.add",
		Fn:          CheckNetAdmin,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-net-raw", Category: "access-control",
		Description: "Verifies containers do not have NET_RAW capability",
		Remediation: "Remove NET_RAW from securityContext.capabilities.add",
		Fn:          CheckNetRaw,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-ipc-lock", Category: "access-control",
		Description: "Verifies containers do not have IPC_LOCK capability",
		Remediation: "Remove IPC_LOCK from securityContext.capabilities.add",
		Fn:          CheckIPCLock,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-bpf", Category: "access-control",
		Description: "Verifies containers do not have BPF capability",
		Remediation: "Remove BPF from securityContext.capabilities.add",
		Fn:          CheckBPF,
	})

	// Host checks
	checks.Register(checks.CheckInfo{
		Name: "access-control-host-network", Category: "access-control",
		Description: "Verifies pods do not use HostNetwork",
		Remediation: "Set spec.hostNetwork to false",
		Fn:          CheckHostNetwork,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-host-path", Category: "access-control",
		Description: "Verifies pods do not use HostPath volumes",
		Remediation: "Remove HostPath volumes from pod spec",
		Fn:          CheckHostPath,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-host-ipc", Category: "access-control",
		Description: "Verifies pods do not use HostIPC",
		Remediation: "Set spec.hostIPC to false",
		Fn:          CheckHostIPC,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-host-pid", Category: "access-control",
		Description: "Verifies pods do not use HostPID",
		Remediation: "Set spec.hostPID to false",
		Fn:          CheckHostPID,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-container-host-port", Category: "access-control",
		Description: "Verifies containers do not use HostPort",
		Remediation: "Remove hostPort from container port definitions",
		Fn:          CheckContainerHostPort,
	})

	// Security context checks
	checks.Register(checks.CheckInfo{
		Name: "access-control-non-root-user", Category: "access-control",
		Description: "Verifies containers set runAsNonRoot to true",
		Remediation: "Set securityContext.runAsNonRoot to true",
		Fn:          CheckNonRootUser,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-privilege-escalation", Category: "access-control",
		Description: "Verifies containers set allowPrivilegeEscalation to false",
		Remediation: "Set securityContext.allowPrivilegeEscalation to false",
		Fn:          CheckPrivilegeEscalation,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-read-only-filesystem", Category: "access-control",
		Description: "Verifies containers set readOnlyRootFilesystem to true",
		Remediation: "Set securityContext.readOnlyRootFilesystem to true",
		Fn:          CheckReadOnlyFilesystem,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-1337-uid", Category: "access-control",
		Description: "Verifies containers do not run as UID 1337 (reserved by Istio)",
		Remediation: "Use a UID other than 1337",
		Fn:          Check1337UID,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-security-context", Category: "access-control",
		Description: "Categorizes container security context (SCC classification)",
		Remediation: "Ensure containers do not require privileged SCC",
		Fn:          CheckSecurityContext,
	})

	// RBAC checks
	checks.Register(checks.CheckInfo{
		Name: "access-control-service-account", Category: "access-control",
		Description: "Verifies pods do not use the default service account",
		Remediation: "Create and assign a dedicated service account",
		Fn:          CheckServiceAccount,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-role-bindings", Category: "access-control",
		Description: "Verifies RoleBindings do not reference non-target namespace ServiceAccounts",
		Remediation: "Ensure RoleBindings only reference ServiceAccounts from target namespaces",
		Fn:          CheckRoleBindings,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-cluster-role-bindings", Category: "access-control",
		Description: "Verifies pod ServiceAccounts are not bound to ClusterRoleBindings",
		Remediation: "Use namespace-scoped RoleBindings instead of ClusterRoleBindings",
		Fn:          CheckClusterRoleBindings,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-automount-token", Category: "access-control",
		Description: "Verifies pods do not automount service account tokens",
		Remediation: "Set automountServiceAccountToken to false on the pod or service account",
		Fn:          CheckAutomountToken,
	})

	// Service checks
	checks.Register(checks.CheckInfo{
		Name: "access-control-nodeport-service", Category: "access-control",
		Description: "Verifies services do not use NodePort type",
		Remediation: "Use ClusterIP or LoadBalancer service type instead",
		Fn:          CheckNodePortService,
	})

	// Probe-based checks
	checks.Register(checks.CheckInfo{
		Name: "access-control-one-process", Category: "access-control",
		Description: "Verifies each container runs only one process",
		Remediation: "Ensure each container has a single main process",
		Fn:          CheckOneProcess,
	})
	checks.Register(checks.CheckInfo{
		Name: "access-control-no-sshd", Category: "access-control",
		Description: "Verifies no SSH daemons are running in containers",
		Remediation: "Remove SSH daemon from container images",
		Fn:          CheckNoSSHD,
	})
}
