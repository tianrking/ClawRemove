package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/tianrking/ClawRemove/internal/model"
)

// PrintEnvironment prints a full environment report.
func PrintEnvironment(w io.Writer, report model.EnvironmentReport, jsonMode bool) error {
	if jsonMode {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}

	lines := []string{
		"Agent Environment Report",
		"========================",
		"",
		"Platform: " + report.Platform,
		"Architecture: " + report.Host.Arch,
		"",
		"AI Runtime",
		"----------",
		report.Runtime.Summary,
	}
	for _, rt := range report.Runtime.Detected {
		status := "stopped"
		if rt.Running {
			status = "running"
		}
		line := fmt.Sprintf("  %s (%s)", rt.Name, status)
		if rt.Version != "" {
			line += " v" + rt.Version
		}
		if rt.Port > 0 {
			line += fmt.Sprintf(" port:%d", rt.Port)
		}
		lines = append(lines, line)
	}

	lines = append(lines, "",
		"Agent Tools",
		"-----------",
		report.Agents.Summary)
	if len(report.Agents.Applications) > 0 {
		lines = append(lines, "  Applications:")
		for _, app := range report.Agents.Applications {
			line := fmt.Sprintf("    %s at %s", app.Name, app.Path)
			lines = append(lines, line)
		}
	}
	if len(report.Agents.Frameworks) > 0 {
		lines = append(lines, "  Frameworks:")
		for _, fw := range report.Agents.Frameworks {
			line := fmt.Sprintf("    %s (%s)", fw.Name, fw.Manager)
			if fw.Version != "" {
				line += " v" + fw.Version
			}
			lines = append(lines, line)
		}
	}

	lines = append(lines, "",
		"AI Artifacts",
		"------------",
		report.Artifacts.Summary)
	if len(report.Artifacts.Models) > 0 {
		lines = append(lines, "  Models:")
		for _, m := range report.Artifacts.Models {
			lines = append(lines, fmt.Sprintf("    %s: %s at %s", m.Name, formatSize(m.Size), m.Path))
		}
	}
	if len(report.Artifacts.Caches) > 0 {
		lines = append(lines, "  Caches:")
		for _, c := range report.Artifacts.Caches {
			lines = append(lines, fmt.Sprintf("    %s: %s at %s", c.Name, formatSize(c.Size), c.Path))
		}
	}

	lines = append(lines, "",
		"Security",
		"--------",
		report.Security.Summary)
	for _, f := range report.Security.Findings {
		severity := strings.ToUpper(f.Severity)
		if severity == "HIGH" {
			severity = "⚠️ HIGH"
		}
		lines = append(lines, fmt.Sprintf("  [%s] %s: %s", severity, f.Provider, f.Location))
		if f.Remediation != "" {
			lines = append(lines, fmt.Sprintf("    → %s", f.Remediation))
		}
	}

	lines = append(lines, "",
		"Hygiene",
		"-------",
		report.Hygiene.Summary)
	if len(report.Hygiene.Recommendations) > 0 {
		lines = append(lines, "  Recommendations:")
		for _, rec := range report.Hygiene.Recommendations {
			lines = append(lines, "    - "+rec)
		}
	}

	_, err := io.WriteString(w, strings.Join(lines, "\n")+"\n")
	return err
}

// PrintInventory prints only the inventory section.
func PrintInventory(w io.Writer, report model.EnvironmentReport, jsonMode bool) error {
	if jsonMode {
		inventory := struct {
			Runtime   model.RuntimeSection   `json:"runtime"`
			Agents    model.AgentsSection    `json:"agents"`
			Artifacts model.ArtifactsSection `json:"artifacts"`
		}{
			Runtime:   report.Runtime,
			Agents:    report.Agents,
			Artifacts: report.Artifacts,
		}
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(inventory)
	}

	lines := []string{
		"AI Inventory",
		"============",
		"",
		"Runtime: " + report.Runtime.Summary,
	}
	for _, rt := range report.Runtime.Detected {
		status := "stopped"
		if rt.Running {
			status = "running"
		}
		lines = append(lines, fmt.Sprintf("  - %s [%s]", rt.Name, status))
	}

	lines = append(lines, "", "Agents: "+report.Agents.Summary)
	for _, app := range report.Agents.Applications {
		lines = append(lines, fmt.Sprintf("  - %s: %s", app.Name, app.Path))
	}
	for _, fw := range report.Agents.Frameworks {
		lines = append(lines, fmt.Sprintf("  - %s (%s)", fw.Name, fw.Manager))
	}

	lines = append(lines, "", "Artifacts: "+report.Artifacts.Summary)
	for _, m := range report.Artifacts.Models {
		lines = append(lines, fmt.Sprintf("  - %s: %s", m.Name, formatSize(m.Size)))
	}

	_, err := io.WriteString(w, strings.Join(lines, "\n")+"\n")
	return err
}

// PrintSecurity prints only the security section.
func PrintSecurity(w io.Writer, report model.EnvironmentReport, jsonMode bool) error {
	if jsonMode {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report.Security)
	}

	lines := []string{
		"AI Security Audit",
		"=================",
		"",
		"Summary: " + report.Security.Summary,
	}
	if report.Security.HighRisk > 0 {
		lines = append(lines, fmt.Sprintf("High Risk Issues: %d", report.Security.HighRisk))
	}

	if len(report.Security.Findings) > 0 {
		lines = append(lines, "", "Findings:")
		for _, f := range report.Security.Findings {
			severity := strings.ToUpper(f.Severity)
			if severity == "HIGH" {
				severity = "⚠️ HIGH"
			}
			lines = append(lines, fmt.Sprintf("  [%s] %s", severity, f.Type))
			lines = append(lines, fmt.Sprintf("    Provider: %s", f.Provider))
			lines = append(lines, fmt.Sprintf("    Location: %s", f.Location))
			if f.Line > 0 {
				lines = append(lines, fmt.Sprintf("    Line: %d", f.Line))
			}
			lines = append(lines, fmt.Sprintf("    Fix: %s", f.Remediation))
		}
	} else {
		lines = append(lines, "", "No security issues found.")
	}

	_, err := io.WriteString(w, strings.Join(lines, "\n")+"\n")
	return err
}

// PrintHygiene prints only the hygiene section.
func PrintHygiene(w io.Writer, report model.EnvironmentReport, jsonMode bool) error {
	if jsonMode {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report.Hygiene)
	}

	lines := []string{
		"AI Storage Hygiene",
		"==================",
		"",
		"Summary: " + report.Hygiene.Summary,
		"",
		"Storage Usage:",
		fmt.Sprintf("  Models:    %s", formatSize(report.Hygiene.ModelsSize)),
		fmt.Sprintf("  Cache:     %s", formatSize(report.Hygiene.CacheSize)),
		fmt.Sprintf("  Vector DB: %s", formatSize(report.Hygiene.VectorDBSize)),
		fmt.Sprintf("  Logs:      %s", formatSize(report.Hygiene.LogSize)),
		"  -------------------",
		fmt.Sprintf("  Total:     %s", formatSize(report.Hygiene.TotalSize)),
	}

	if len(report.Hygiene.Recommendations) > 0 {
		lines = append(lines, "", "Recommendations:")
		for _, rec := range report.Hygiene.Recommendations {
			lines = append(lines, "  - "+rec)
		}
	}

	_, err := io.WriteString(w, strings.Join(lines, "\n")+"\n")
	return err
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1fTB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

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

func PrintProducts(w io.Writer, providers []model.ProductFacts, jsonMode bool) error {
	if jsonMode {
		_, err := io.WriteString(w, "[")
		if err != nil {
			return err
		}
		for i, p := range providers {
			if i > 0 {
				if _, err := io.WriteString(w, ","); err != nil {
					return err
				}
			}
			if _, err := io.WriteString(w, fmt.Sprintf(`{"id":"%s","displayName":"%s"}`, p.ID, p.DisplayName)); err != nil {
				return err
			}
		}
		_, err = io.WriteString(w, "]\n")
		return err
	}
	for _, p := range providers {
		if _, err := io.WriteString(w, p.ID+"\t"+p.DisplayName+"\n"); err != nil {
			return err
		}
	}
	return nil
}

func ConfirmApply(reader io.Reader, stdout io.Writer, stderr io.Writer, product string, report model.Report) (bool, error) {
	_, _ = io.WriteString(stdout, "\nSafety check:\n")
	_, _ = io.WriteString(stdout, fmt.Sprintf("- Confirmed residuals: %d\n", len(report.Verify.Confirmed)))
	_, _ = io.WriteString(stdout, fmt.Sprintf("- Investigate residuals: %d\n", len(report.Verify.Investigate)))
	_, _ = io.WriteString(stdout, fmt.Sprintf("- Planned actions: %d\n", len(report.Plan.Actions)))
	_, _ = io.WriteString(stdout, fmt.Sprintf("Type REMOVE %s to continue: ", strings.ToUpper(product)))

	br := bufio.NewReader(reader)
	line, err := br.ReadString('\n')
	if err != nil && !strings.Contains(err.Error(), "EOF") {
		return false, err
	}
	expected := "REMOVE " + strings.ToUpper(product)
	if strings.TrimSpace(line) != expected {
		_, _ = io.WriteString(stderr, "Confirmation phrase did not match. No removal actions were executed.\n")
		return false, nil
	}
	return true, nil
}
