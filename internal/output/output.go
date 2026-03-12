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
		fmt.Sprintf("State dirs: %d", len(report.Discovery.StateDirs)),
		fmt.Sprintf("Workspace dirs: %d", len(report.Discovery.WorkspaceDirs)),
		fmt.Sprintf("Services: %d", len(report.Discovery.Services)),
		fmt.Sprintf("Packages: %d", len(report.Discovery.Packages)),
		fmt.Sprintf("Processes: %d", len(report.Discovery.Processes)),
		fmt.Sprintf("Containers: %d", len(report.Discovery.Containers)),
		fmt.Sprintf("Images: %d", len(report.Discovery.Images)),
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
	_, err := io.WriteString(w, strings.Join(lines, "\n")+"\n")
	return err
}
