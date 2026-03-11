package platform

import (
	"context"
	"fmt"
	"strings"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// CheckHugepages verifies hugepage configuration on nodes (probe-based).
func CheckHugepages(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if resources.ProbeExecutor == nil || len(resources.ProbePods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "Probe pods not available"
		return result
	}

	ctx := context.Background()
	var count int
	for nodeName, probePod := range resources.ProbePods {
		// Check for hugepage kernel boot parameters
		stdout, _, err := resources.ProbeExecutor.ExecCommand(ctx, probePod, "cat /host/proc/cmdline")
		if err != nil {
			continue
		}
		if strings.Contains(stdout, "hugepagesz") || strings.Contains(stdout, "hugepages=") {
			// Hugepages configured via kernel boot params — check consistency with sysfs
			sysStdout, _, sysErr := resources.ProbeExecutor.ExecCommand(ctx, probePod,
				"cat /host/sys/kernel/mm/hugepages/hugepages-2048kB/nr_hugepages 2>/dev/null")
			if sysErr != nil {
				continue
			}
			nrHugepages := strings.TrimSpace(sysStdout)
			if nrHugepages == "0" {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Node", Name: nodeName, Namespace: "",
					Compliant: false,
					Message:   "Hugepages configured in boot params but nr_hugepages is 0",
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d node(s) have misconfigured hugepages", count)
	}
	return result
}
