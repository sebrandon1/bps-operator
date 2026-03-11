package accesscontrol

import (
	"context"
	"fmt"
	"strings"

	"github.com/sebrandon1/bps-operator/internal/checks"
)

// CheckOneProcess verifies each container runs only one process (probe-based).
func CheckOneProcess(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if resources.ProbeExecutor == nil || len(resources.ProbePods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "Probe pods not available"
		return result
	}

	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	ctx := context.Background()
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		probePod, ok := resources.ProbePods[pod.Spec.NodeName]
		if !ok || probePod == nil {
			continue
		}

		for _, container := range pod.Spec.Containers {
			// Use lsns to count namespaces — each process in its own PID namespace
			cmd := fmt.Sprintf("lsns -t pid -o PID,COMMAND --no-headings | grep -c '%s'", container.Name)
			stdout, _, err := resources.ProbeExecutor.ExecCommand(ctx, probePod, cmd)
			if err != nil {
				continue
			}
			stdout = strings.TrimSpace(stdout)
			if stdout != "" && stdout != "0" && stdout != "1" {
				count++
				result.Details = append(result.Details, checks.ResourceDetail{
					Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
					Compliant: false,
					Message:   fmt.Sprintf("Container %q has multiple processes", container.Name),
				})
			}
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d container(s) have multiple processes", count)
	}
	return result
}

// CheckNoSSHD verifies no SSH daemons are running (probe-based).
func CheckNoSSHD(resources *checks.DiscoveredResources) checks.CheckResult {
	result := checks.CheckResult{ComplianceStatus: "Compliant"}
	if resources.ProbeExecutor == nil || len(resources.ProbePods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "Probe pods not available"
		return result
	}

	if len(resources.Pods) == 0 {
		result.ComplianceStatus = "Skipped"
		result.Reason = "No pods found"
		return result
	}

	var count int
	ctx := context.Background()
	for i := range resources.Pods {
		pod := &resources.Pods[i]
		probePod, ok := resources.ProbePods[pod.Spec.NodeName]
		if !ok || probePod == nil {
			continue
		}

		cmd := fmt.Sprintf("nsenter --target $(crictl inspect $(crictl ps --name %s -q 2>/dev/null | head -1) 2>/dev/null | jq -r '.info.pid' 2>/dev/null) --mount --pid -- pgrep -x sshd 2>/dev/null", pod.Name)
		stdout, _, err := resources.ProbeExecutor.ExecCommand(ctx, probePod, cmd)
		if err != nil {
			continue
		}
		if strings.TrimSpace(stdout) != "" {
			count++
			result.Details = append(result.Details, checks.ResourceDetail{
				Kind: "Pod", Name: pod.Name, Namespace: pod.Namespace,
				Compliant: false, Message: "SSH daemon found running",
			})
		}
	}
	if count > 0 {
		result.ComplianceStatus = "NonCompliant"
		result.Reason = fmt.Sprintf("%d pod(s) have SSH daemons running", count)
	}
	return result
}
