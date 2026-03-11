package platform

import (
	"context"
	"fmt"
	"strings"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// grubKernelArgs are kernel boot parameters that should not be modified.
var grubKernelArgs = []string{
	"hugepagesz",
	"hugepages",
	"isolcpus",
	"rcu_nocbs",
	"rcu_nocb_poll",
	"nohz_full",
	"tuned.non_isolcpus",
}

// CheckBootParams verifies no non-standard kernel boot parameters are set (probe-based).
func CheckBootParams(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if resources.ProbeExecutor == nil || len(resources.ProbePods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "Probe pods not available"
		return result
	}

	ctx := context.Background()
	var count int
	for nodeName, probePod := range resources.ProbePods {
		stdout, _, err := resources.ProbeExecutor.ExecCommand(ctx, probePod, "cat /host/proc/cmdline")
		if err != nil {
			continue
		}
		for _, arg := range grubKernelArgs {
			if strings.Contains(stdout, arg) {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Node", Name: nodeName, Namespace: "",
					Compliant: false,
					Message:   fmt.Sprintf("Boot parameter %q found in kernel cmdline", arg),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d non-standard boot parameter(s) found", count)
	}
	return result
}
