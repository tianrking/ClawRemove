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
		"AI Environment Report",
		"─────────────────────────",
	}

	// AI Runtime section
	if len(report.Runtime.Detected) > 0 {
		lines = append(lines, "",
			"AI Runtime")
		for _, rt := range report.Runtime.Detected {
			status := "stopped"
			if rt.Running {
				status = "running"
			}
			line := fmt.Sprintf("  - %s (%s)", rt.Name, status)
			if rt.Port > 0 {
				line += fmt.Sprintf(" port:%d", rt.Port)
			}
			lines = append(lines, line)
			if rt.Path != "" {
				lines = append(lines, fmt.Sprintf("      %s", rt.Path))
			}
		}
	}

	// AI Tools section
	if len(report.Agents.Applications) > 0 {
		lines = append(lines, "",
			"AI Tools")
		for _, app := range report.Agents.Applications {
			lines = append(lines, fmt.Sprintf("  - %s", app.Name))
			if app.Path != "" {
				lines = append(lines, fmt.Sprintf("      %s", app.Path))
			}
		}
	}

	// Models section
	if len(report.Artifacts.Models) > 0 {
		lines = append(lines, "",
			"Models")
		for _, m := range report.Artifacts.Models {
			lines = append(lines, fmt.Sprintf("  - %s: %s", m.Name, formatSize(m.Size)))
			if m.Path != "" {
				lines = append(lines, fmt.Sprintf("      %s", m.Path))
			}
		}
	}

	// Caches section
	if len(report.Artifacts.Caches) > 0 {
		lines = append(lines, "",
			"Caches")
		for _, c := range report.Artifacts.Caches {
			lines = append(lines, fmt.Sprintf("  - %s: %s", c.Name, formatSize(c.Size)))
			if c.Path != "" {
				lines = append(lines, fmt.Sprintf("      %s", c.Path))
			}
		}
	}

	// Vector DBs section
	if len(report.Artifacts.VectorDBs) > 0 {
		lines = append(lines, "",
			"Vector Databases")
		for _, v := range report.Artifacts.VectorDBs {
			status := ""
			if v.Size > 0 {
				status = fmt.Sprintf(": %s", formatSize(v.Size))
			}
			lines = append(lines, fmt.Sprintf("  - %s%s", v.Name, status))
			if v.Path != "" {
				lines = append(lines, fmt.Sprintf("      %s", v.Path))
			}
		}
	}

	// Security findings
	if len(report.Security.Findings) > 0 {
		lines = append(lines, "",
			"Security Findings")
		for _, f := range report.Security.Findings {
			icon := "⚠️"
			if f.Severity == "high" {
				icon = "🔴"
			}
			lines = append(lines, fmt.Sprintf("  %s %s found in:", icon, f.Provider))
			lines = append(lines, fmt.Sprintf("      %s", f.Location))
		}
	}

	// Total AI Storage - the killer feature!
	lines = append(lines, "",
		"─────────────────────────",
		fmt.Sprintf("Total AI Storage: %s", formatSize(report.Hygiene.TotalSize)))

	// Recommendations
	if len(report.Hygiene.Recommendations) > 0 {
		lines = append(lines, "",
			"Recommendations:")
		for _, rec := range report.Hygiene.Recommendations {
			lines = append(lines, fmt.Sprintf("  • %s", rec))
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
		"─────────────────────────",
	}

	// Runtime
	if len(report.Runtime.Detected) > 0 {
		lines = append(lines, "",
			"AI Runtime")
		for _, rt := range report.Runtime.Detected {
			status := "stopped"
			if rt.Running {
				status = "running"
			}
			lines = append(lines, fmt.Sprintf("  - %s (%s)", rt.Name, status))
			if rt.Path != "" {
				lines = append(lines, fmt.Sprintf("      %s", rt.Path))
			}
		}
	}

	// Agents
	if len(report.Agents.Applications) > 0 {
		lines = append(lines, "",
			"AI Tools")
		for _, app := range report.Agents.Applications {
			lines = append(lines, fmt.Sprintf("  - %s", app.Name))
			if app.Path != "" {
				lines = append(lines, fmt.Sprintf("      %s", app.Path))
			}
		}
	}

	// Frameworks
	if len(report.Agents.Frameworks) > 0 {
		lines = append(lines, "",
			"Frameworks")
		for _, fw := range report.Agents.Frameworks {
			lines = append(lines, fmt.Sprintf("  - %s (%s)", fw.Name, fw.Manager))
		}
	}

	// Artifacts
	if len(report.Artifacts.Models) > 0 || len(report.Artifacts.Caches) > 0 {
		lines = append(lines, "",
			"Artifacts")
		for _, m := range report.Artifacts.Models {
			lines = append(lines, fmt.Sprintf("  - %s: %s", m.Name, formatSize(m.Size)))
			if m.Path != "" {
				lines = append(lines, fmt.Sprintf("      %s", m.Path))
			}
		}
		for _, c := range report.Artifacts.Caches {
			lines = append(lines, fmt.Sprintf("  - %s: %s", c.Name, formatSize(c.Size)))
			if c.Path != "" {
				lines = append(lines, fmt.Sprintf("      %s", c.Path))
			}
		}
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
		"─────────────────────────",
	}

	if len(report.Security.Findings) > 0 {
		for _, f := range report.Security.Findings {
			icon := "⚠️"
			if f.Severity == "high" {
				icon = "🔴"
			}
			lines = append(lines, fmt.Sprintf("  %s %s", icon, f.Type))
			lines = append(lines, fmt.Sprintf("      Provider: %s", f.Provider))
			lines = append(lines, fmt.Sprintf("      Location: %s", f.Location))
			if f.Line > 0 {
				lines = append(lines, fmt.Sprintf("      Line: %d", f.Line))
			}
		}
		lines = append(lines, "",
			"─────────────────────────")
		if report.Security.HighRisk > 0 {
			lines = append(lines, fmt.Sprintf("High Risk Issues: %d", report.Security.HighRisk))
		}
		lines = append(lines, fmt.Sprintf("Total Issues: %d", len(report.Security.Findings)))
	} else {
		lines = append(lines, "",
			"No security issues found.")
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
		"AI Storage Usage",
		"─────────────────────────",
	}
	if report.Hygiene.ModelsSize > 0 {
		lines = append(lines, fmt.Sprintf("  Models:     %s", formatSize(report.Hygiene.ModelsSize)))
	}
	if report.Hygiene.CacheSize > 0 {
		lines = append(lines, fmt.Sprintf("  Cache:      %s", formatSize(report.Hygiene.CacheSize)))
	}
	if report.Hygiene.VectorDBSize > 0 {
		lines = append(lines, fmt.Sprintf("  Vector DB:  %s", formatSize(report.Hygiene.VectorDBSize)))
	}
	if report.Hygiene.LogSize > 0 {
		lines = append(lines, fmt.Sprintf("  Logs:       %s", formatSize(report.Hygiene.LogSize)))
	}

	lines = append(lines, "─────────────────────────")
	lines = append(lines, fmt.Sprintf("  Total:      %s", formatSize(report.Hygiene.TotalSize)))

	if len(report.Hygiene.Recommendations) > 0 {
		lines = append(lines, "",
			"Recommendations:")
		for _, rec := range report.Hygiene.Recommendations {
			lines = append(lines, fmt.Sprintf("  • %s", rec))
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

// PrintCleanup prints a cleanup scan report.
func PrintCleanup(w io.Writer, report model.CleanupReport, jsonMode bool) error {
	if jsonMode {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}

	lines := []string{
		"AI Cleanup Scan",
		"─────────────────────────",
	}

	if len(report.Candidates) == 0 {
		lines = append(lines, "",
			"No cleanup candidates found.")
	} else {
		// Group by category
		categories := map[string][]model.CleanupCandidate{}
		categoryOrder := []string{"model_version", "orphaned_cache", "unused_vectordb", "log_rotation"}
		for _, c := range report.Candidates {
			categories[c.Category] = append(categories[c.Category], c)
		}

		categoryNames := map[string]string{
			"model_version":    "📦 Old Model Versions",
			"orphaned_cache":   "🗑️  Orphaned Caches",
			"unused_vectordb":  "🗄️  Unused Vector Databases",
			"log_rotation":     "📝 Log Files",
		}

		for _, cat := range categoryOrder {
			items, ok := categories[cat]
			if !ok || len(items) == 0 {
				continue
			}

			catName := categoryNames[cat]
			if catName == "" {
				catName = cat
			}

			var catSize int64
			for _, item := range items {
				catSize += item.Size
			}

			lines = append(lines, "",
				fmt.Sprintf("%s (%s)", catName, formatSize(catSize)))

			for _, item := range items {
				risk := ""
				if item.Risk == "medium" {
					risk = " ⚠️"
				} else if item.Risk == "high" {
					risk = " 🔴"
				}
				lines = append(lines, fmt.Sprintf("  - %s%s", item.Reason, risk))
				lines = append(lines, fmt.Sprintf("      %s (%s)", item.Path, formatSize(item.Size)))
			}
		}

		lines = append(lines, "",
			"─────────────────────────",
			fmt.Sprintf("Total reclaimable: %s (%d items)", formatSize(report.TotalReclaimable), len(report.Candidates)))
	}

	_, err := io.WriteString(w, strings.Join(lines, "\n")+"\n")
	return err
}

