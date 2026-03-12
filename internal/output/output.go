package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/tianrking/ClawRemove/internal/model"
)

func PrintReport(w io.Writer, report model.Report, jsonMode bool) error {
	if jsonMode {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}

	lines := []string{
		"Claw Remove",
		"Product: " + report.Product,
		"Command: " + report.Command,
		"Platform: " + report.Discovery.Platform,
		"Architecture: " + report.Host.Arch,
		fmt.Sprintf("State dirs: %d", len(report.Discovery.StateDirs)),
		fmt.Sprintf("Workspace dirs: %d", len(report.Discovery.WorkspaceDirs)),
		fmt.Sprintf("Services: %d", len(report.Discovery.Services)),
		fmt.Sprintf("Packages: %d", len(report.Discovery.Packages)),
		fmt.Sprintf("Processes: %d", len(report.Discovery.Processes)),
		fmt.Sprintf("Containers: %d", len(report.Discovery.Containers)),
		fmt.Sprintf("Images: %d", len(report.Discovery.Images)),
		fmt.Sprintf("Provider skills: %d", len(report.Capabilities.Skills)),
		fmt.Sprintf("Provider tools: %d", len(report.Capabilities.Tools)),
		fmt.Sprintf("Evidence: exact=%d strong=%d heuristic=%d", report.Evidence.Summary.Exact, report.Evidence.Summary.Strong, report.Evidence.Summary.Heuristic),
		fmt.Sprintf("Verified residuals: exact=%d strong=%d heuristic=%d", report.Verify.Summary.Exact, report.Verify.Summary.Strong, report.Verify.Summary.Heuristic),
		fmt.Sprintf("Planned actions: %d", len(report.Plan.Actions)),
	}
	if report.AuditOnly {
		lines = append(lines, "Mode: audit-only")
	} else if report.DryRun {
		lines = append(lines, "Mode: dry-run")
	}
	if len(report.Results) > 0 {
		lines = append(lines, "", "Results:")
		for _, result := range report.Results {
			status := "ok"
			if !result.OK {
				status = "fail"
			} else if result.Skipped {
				status = "skip"
			}
			line := fmt.Sprintf("- [%s] %s :: %s", status, result.Action, result.Target)
			if result.Error != "" {
				line += " :: " + result.Error
			}
			lines = append(lines, line)
		}
	}
	if report.Advice != nil {
		lines = append(lines, "", "Advisor:")
		lines = append(lines, "- Mode: "+report.Advice.Mode)
		lines = append(lines, "- Authority: "+report.Advice.Authority)
		lines = append(lines, "- Summary: "+report.Advice.ThoughtSummary)
		lines = append(lines, "- Message: "+report.Advice.UserMessage)
		for _, rec := range report.Advice.Recommendations {
			line := fmt.Sprintf("- Recommendation: %s :: %s :: risk=%s :: evidence=%s", rec.Kind, rec.Target, rec.Risk, rec.Evidence)
			lines = append(lines, line)
		}
	}
	if len(report.Capabilities.Skills) > 0 || len(report.Capabilities.Tools) > 0 {
		lines = append(lines, "", "Provider Capabilities:")
		for _, skill := range report.Capabilities.Skills {
			lines = append(lines, fmt.Sprintf("- Skill: %s :: %s", skill.ID, skill.Name))
		}
		for _, tool := range report.Capabilities.Tools {
			lines = append(lines, fmt.Sprintf("- Tool: %s :: readOnly=%t", tool.ID, tool.ReadOnly))
		}
	}
	if report.Verify.Verified {
		lines = append(lines, "", "Verification:")
		for _, residual := range report.Verify.Confirmed {
			lines = append(lines, fmt.Sprintf("- Confirmed: %s :: %s :: evidence=%s", residual.Kind, residual.Target, residual.Evidence))
		}
		for _, residual := range report.Verify.Investigate {
			lines = append(lines, fmt.Sprintf("- Investigate: %s :: %s :: evidence=%s", residual.Kind, residual.Target, residual.Evidence))
		}
	}
	_, err := io.WriteString(w, strings.Join(lines, "\n")+"\n")
	return err
}
