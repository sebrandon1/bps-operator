package platform

import (
	"context"
	"fmt"
	"strings"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// mcSysctls are sysctl settings that should not be modified outside of MachineConfig.
var mcSysctls = []string{
	"net.ipv4.conf.all.accept_redirects",
	"net.ipv6.conf.all.accept_redirects",
	"net.ipv4.conf.all.secure_redirects",
	"kernel.core_pattern",
}

// CheckSysctl verifies sysctl settings are not modified outside of MachineConfig (probe-based).
func CheckSysctl(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if resources.ProbeExecutor == nil || len(resources.ProbePods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "Probe pods not available"
		return result
	}

	ctx := context.Background()
	var count int
	for nodeName, probePod := range resources.ProbePods {
		for _, sysctl := range mcSysctls {
			cmd := fmt.Sprintf("chroot /host sysctl -n %s 2>/dev/null", sysctl)
			stdout, _, err := resources.ProbeExecutor.ExecCommand(ctx, probePod, cmd)
			if err != nil {
				continue
			}
			val := strings.TrimSpace(stdout)
			if isNonDefaultSysctl(sysctl, val) {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Node", Name: nodeName, Namespace: "",
					Compliant: false,
					Message:   fmt.Sprintf("Sysctl %s has non-default value %q", sysctl, val),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d non-default sysctl setting(s) found", count)
	}
	return result
}

func isNonDefaultSysctl(name, value string) bool {
	defaults := map[string]string{
		"net.ipv4.conf.all.accept_redirects":  "0",
		"net.ipv6.conf.all.accept_redirects":  "0",
		"net.ipv4.conf.all.secure_redirects":  "1",
		"kernel.core_pattern":                 "|/usr/lib/systemd/systemd-coredump %P %u %g %s %t %c %h",
	}
	if expected, ok := defaults[name]; ok {
		return value != expected
	}
	return false
}
