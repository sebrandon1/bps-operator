package platform

import (
	"context"
	"fmt"
	"strings"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// CheckTainted verifies the kernel is not tainted (probe-based).
func CheckTainted(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if resources.ProbeExecutor == nil || len(resources.ProbePods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "Probe pods not available"
		return result
	}

	ctx := context.Background()
	var count int
	for nodeName, probePod := range resources.ProbePods {
		stdout, _, err := resources.ProbeExecutor.ExecCommand(ctx, probePod, "cat /host/proc/sys/kernel/tainted")
		if err != nil {
			continue
		}
		val := strings.TrimSpace(stdout)
		if val != "0" {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Node", Name: nodeName, Namespace: "",
				Compliant: false,
				Message:   fmt.Sprintf("Kernel taint value is %s (expected 0)", val),
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d node(s) have tainted kernels", count)
	}
	return result
}
