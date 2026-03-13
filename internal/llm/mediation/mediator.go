package mediation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/platform"
	"github.com/tianrking/ClawRemove/internal/system"
	"github.com/tianrking/ClawRemove/internal/tools"
)

type Mediator struct {
	runner        system.Runner
	adapter       platform.Adapter
	providerTools map[string]tools.Tool
}

func New(runner system.Runner, adapter platform.Adapter, providerTools []tools.Tool) Mediator {
	toolMap := make(map[string]tools.Tool, len(providerTools))
	for _, t := range providerTools {
		toolMap[t.Info().ID] = t
	}
	return Mediator{runner: runner, adapter: adapter, providerTools: toolMap}
}

func (m Mediator) ExecuteTool(ctx context.Context, report model.Report, tool string, input map[string]any) (any, error) {
	limit := 10
	if raw, ok := input["limit"]; ok {
		switch v := raw.(type) {
		case float64:
			if v > 0 {
				limit = int(v)
			}
		case int:
			if v > 0 {
				limit = v
			}
		}
	}

	if t, ok := m.providerTools[tool]; ok {
		res, err := t.Execute(ctx, report, input)
		if err != nil {
			return nil, err
		}
		if res != nil {
			return res, nil
		}
	}

	switch tool {
	case "summary":
		return toolSummary(report), nil
	case "verification":
		return report.Verify, nil
	case "state_dirs":
		return sliceResult("stateDirs", report.Discovery.StateDirs, limit), nil
	case "workspace_dirs":
		return sliceResult("workspaceDirs", report.Discovery.WorkspaceDirs, limit), nil
	case "services":
		return sliceResult("services", report.Discovery.Services, limit), nil
	case "packages":
		return sliceResult("packages", report.Discovery.Packages, limit), nil
	case "processes":
		return sliceResult("processes", report.Discovery.Processes, limit), nil
	case "containers":
		return map[string]any{
			"containers": truncateSlice(report.Discovery.Containers, limit),
			"images":     truncateSlice(report.Discovery.Images, limit),
		}, nil
	case "plan_actions":
		return sliceResult("planActions", report.Plan.Actions, limit), nil
	case "deep_analysis":
		return m.deepAnalysis(report), nil
	case "registry_probe":
		return m.registryProbe(report, input)
	case "env_probe":
		return m.envProbe(report, input)
	case "hosts_probe":
		return m.hostsProbe(report, input)
	case "autostart_probe":
		return m.autostartProbe(report, input)
	case "path_probe":
		return m.pathProbe(report, input)
	case "shell_profile_probe":
		return m.shellProfileProbe(report, input)
	case "service_probe":
		return m.serviceProbe(report, input)
	case "package_probe":
		return m.packageProbe(report, input)
	case "process_probe":
		return m.processProbe(report, input)
	default:
		return nil, fmt.Errorf("unsupported tool: %s", tool)
	}
}

func (m Mediator) pathProbe(report model.Report, input map[string]any) (any, error) {
	target, err := stringInput(input, "target")
	if err != nil {
		return nil, err
	}
	if !allowedPathTarget(report, target) {
		return nil, fmt.Errorf("path target is not present in the report")
	}
	info, err := os.Stat(target)
	if err != nil {
		return map[string]any{"target": target, "exists": false, "error": err.Error()}, nil
	}
	result := map[string]any{
		"target": target,
		"exists": true,
		"isDir":  info.IsDir(),
		"size":   info.Size(),
		"mode":   info.Mode().String(),
		"base":   filepath.Base(target),
	}
	if info.IsDir() {
		entries, _ := os.ReadDir(target)
		var names []string
		for i, entry := range entries {
			if i >= 20 {
				break
			}
			names = append(names, entry.Name())
		}
		result["entries"] = names
	}
	return result, nil
}

func (m Mediator) shellProfileProbe(report model.Report, input map[string]any) (any, error) {
	target, err := stringInput(input, "target")
	if err != nil {
		return nil, err
	}
	if !contains(report.Discovery.ShellProfiles, target) {
		return nil, fmt.Errorf("shell profile target is not present in the report")
	}
	raw, err := os.ReadFile(target)
	if err != nil {
		return map[string]any{"target": target, "exists": false, "error": err.Error()}, nil
	}
	var matches []string
	for _, line := range strings.Split(string(raw), "\n") {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "openclaw") || strings.Contains(lower, "clawdbot") || strings.Contains(lower, "moltbot") || strings.Contains(lower, "# openclaw completion") {
			matches = append(matches, line)
			if len(matches) >= 20 {
				break
			}
		}
	}
	return map[string]any{
		"target":  target,
		"exists":  true,
		"matches": matches,
		"count":   len(matches),
	}, nil
}

func (m Mediator) serviceProbe(report model.Report, input map[string]any) (any, error) {
	target, err := stringInput(input, "target")
	if err != nil {
		return nil, err
	}
	for _, svc := range report.Discovery.Services {
		if svc.Name != target && svc.Path != target {
			continue
		}
		result := map[string]any{
			"platform": svc.Platform,
			"scope":    svc.Scope,
			"name":     svc.Name,
			"path":     svc.Path,
		}
		if svc.Path != "" {
			raw, err := os.ReadFile(svc.Path)
			if err == nil {
				lines := strings.Split(string(raw), "\n")
				if len(lines) > 30 {
					lines = lines[:30]
				}
				result["filePreview"] = lines
			}
		}
		if cmd := m.adapter.ServiceStatusCommand(svc, currentUID()); len(cmd) > 0 {
			result["status"] = runReadOnly(m.runner, cmd)
		}
		return result, nil
	}
	return nil, fmt.Errorf("service target is not present in the report")
}

func (m Mediator) packageProbe(report model.Report, input map[string]any) (any, error) {
	target, err := stringInput(input, "target")
	if err != nil {
		return nil, err
	}
	for _, pkg := range report.Discovery.Packages {
		key := pkg.Manager + ":" + pkg.Name
		if key != target {
			continue
		}
		cmd := packageReadOnlyCommand(pkg)
		result := map[string]any{
			"manager": pkg.Manager,
			"name":    pkg.Name,
			"kind":    pkg.Kind,
		}
		if len(cmd) > 0 {
			result["query"] = runReadOnly(m.runner, cmd)
		}
		return result, nil
	}
	return nil, fmt.Errorf("package target is not present in the report")
}

func (m Mediator) processProbe(report model.Report, input map[string]any) (any, error) {
	target, err := stringInput(input, "target")
	if err != nil {
		return nil, err
	}
	for _, proc := range report.Discovery.Processes {
		if proc.Command != target {
			continue
		}
		result := map[string]any{
			"pid":     proc.PID,
			"ppid":    proc.PPID,
			"command": proc.Command,
		}
		if cmd := m.adapter.ProcessStatusCommand(proc.PID); len(cmd) > 0 {
			result["status"] = runReadOnly(m.runner, cmd)
		}
		return result, nil
	}
	return nil, fmt.Errorf("process target is not present in the report")
}

// deepAnalysis provides a comprehensive overview of all discovered artifacts
func (m Mediator) deepAnalysis(report model.Report) map[string]any {
	result := map[string]any{
		"product": report.Product,
		"platform": report.Host.OS,
		"summary": map[string]int{
			"stateDirs":       len(report.Discovery.StateDirs),
			"workspaceDirs":   len(report.Discovery.WorkspaceDirs),
			"services":        len(report.Discovery.Services),
			"packages":        len(report.Discovery.Packages),
			"processes":       len(report.Discovery.Processes),
			"containers":      len(report.Discovery.Containers),
			"registryKeys":    len(report.Discovery.RegistryKeys),
			"envVars":         len(report.Discovery.EnvVars),
			"hostsEntries":    len(report.Discovery.HostsEntries),
			"shellProfiles":   len(report.Discovery.ShellProfiles),
			"crontabLines":    len(report.Discovery.CrontabLines),
			"confirmed":       len(report.Verify.Confirmed),
			"investigate":     len(report.Verify.Investigate),
		},
	}

	// Categorize findings by modification type
	modifications := map[string][]string{
		"filesystem":   {},
		"services":     {},
		"environment":  {},
		"network":      {},
		"autostart":    {},
	}

	// Analyze state directories
	for _, dir := range report.Discovery.StateDirs {
		modifications["filesystem"] = append(modifications["filesystem"], dir)
	}

	// Analyze services
	for _, svc := range report.Discovery.Services {
		modifications["services"] = append(modifications["services"], svc.Name)
		if svc.Scope == "system" || svc.Scope == "user" {
			modifications["autostart"] = append(modifications["autostart"], svc.Name+" ("+svc.Platform+")")
		}
	}

	// Analyze environment variables
	for _, env := range report.Discovery.EnvVars {
		modifications["environment"] = append(modifications["environment"], env.Name+"="+env.Value)
	}

	// Analyze hosts entries
	for _, entry := range report.Discovery.HostsEntries {
		modifications["network"] = append(modifications["network"], entry)
	}

	// Analyze crontab
	for _, line := range report.Discovery.CrontabLines {
		modifications["autostart"] = append(modifications["autostart"], "cron: "+line)
	}

	// Analyze registry keys (Windows)
	for _, reg := range report.Discovery.RegistryKeys {
		modifications["autostart"] = append(modifications["autostart"], reg.RootKey+"\\"+reg.Path)
	}

	result["modifications"] = modifications
	result["analysisNotes"] = []string{
		"Review filesystem modifications for agent configuration and data",
		"Check services for auto-start persistence mechanisms",
		"Examine environment variables for PATH or config modifications",
		"Verify hosts entries for any agent-added domain mappings",
		"Inspect registry keys (Windows) for startup entries",
	}

	return result
}

// registryProbe analyzes Windows registry entries discovered
func (m Mediator) registryProbe(report model.Report, input map[string]any) (any, error) {
	if report.Host.OS != "windows" {
		return map[string]any{"error": "registry probe is only available on Windows", "entries": []any{}}, nil
	}

	limit := 20
	if raw, ok := input["limit"]; ok {
		switch v := raw.(type) {
		case float64:
			if v > 0 && int(v) < 100 {
				limit = int(v)
			}
		case int:
			if v > 0 && v < 100 {
				limit = v
			}
		}
	}

	entries := truncateSlice(report.Discovery.RegistryKeys, limit)
	result := map[string]any{
		"platform": "windows",
		"count":    len(report.Discovery.RegistryKeys),
		"entries":  entries,
		"analysis": []string{},
	}

	// Analyze registry entries for agent modifications
	var analysis []string
	for _, reg := range entries {
		// Check for common auto-start locations
		if strings.Contains(reg.Path, "Run") || strings.Contains(reg.Path, "RunOnce") {
			analysis = append(analysis, fmt.Sprintf("Auto-start entry: %s\\%s", reg.RootKey, reg.Path))
		}
		// Check for uninstall entries
		if strings.Contains(reg.Path, "Uninstall") {
			analysis = append(analysis, fmt.Sprintf("Uninstall entry: %s\\%s", reg.RootKey, reg.Path))
		}
		// Check for app-specific entries
		if reg.Data != "" {
			analysis = append(analysis, fmt.Sprintf("Data entry at %s\\%s: %s", reg.RootKey, reg.Path, reg.Data))
		}
	}
	result["analysis"] = analysis

	return result, nil
}

// envProbe analyzes environment variables discovered
func (m Mediator) envProbe(report model.Report, input map[string]any) (any, error) {
	limit := 20
	if raw, ok := input["limit"]; ok {
		switch v := raw.(type) {
		case float64:
			if v > 0 && int(v) < 100 {
				limit = int(v)
			}
		case int:
			if v > 0 && v < 100 {
				limit = v
			}
		}
	}

	entries := truncateSlice(report.Discovery.EnvVars, limit)
	result := map[string]any{
		"count":   len(report.Discovery.EnvVars),
		"entries": entries,
		"analysis": []string{},
	}

	// Analyze environment variables for agent modifications
	var analysis []string
	for _, env := range entries {
		name := strings.ToUpper(env.Name)
		// Check for PATH modifications
		if strings.Contains(name, "PATH") {
			analysis = append(analysis, fmt.Sprintf("PATH modification: %s", env.Name))
		}
		// Check for agent-specific variables
		if strings.Contains(name, "OPENCLAW") || strings.Contains(name, "CLAW") ||
		   strings.Contains(name, "AGENT") || strings.Contains(name, "BOT") {
			analysis = append(analysis, fmt.Sprintf("Agent-specific variable: %s=%s", env.Name, env.Value))
		}
		// Check for API keys or secrets
		if strings.Contains(name, "API_KEY") || strings.Contains(name, "TOKEN") ||
		   strings.Contains(name, "SECRET") {
			analysis = append(analysis, fmt.Sprintf("Potential secret in env: %s", env.Name))
		}
	}
	result["analysis"] = analysis

	return result, nil
}

// hostsProbe analyzes hosts file entries discovered
func (m Mediator) hostsProbe(report model.Report, input map[string]any) (any, error) {
	limit := 20
	if raw, ok := input["limit"]; ok {
		switch v := raw.(type) {
		case float64:
			if v > 0 && int(v) < 100 {
				limit = int(v)
			}
		case int:
			if v > 0 && v < 100 {
				limit = v
			}
		}
	}

	entries := truncateSlice(report.Discovery.HostsEntries, limit)
	result := map[string]any{
		"count":   len(report.Discovery.HostsEntries),
		"entries": entries,
		"analysis": []string{},
	}

	// Analyze hosts entries for agent modifications
	var analysis []string
	for _, entry := range entries {
		// Check for localhost redirects
		if strings.HasPrefix(entry, "127.0.0.1") || strings.HasPrefix(entry, "::1") {
			analysis = append(analysis, fmt.Sprintf("Localhost redirect: %s", entry))
		}
		// Check for API endpoint modifications
		if strings.Contains(entry, "api.") || strings.Contains(entry, "openai") ||
		   strings.Contains(entry, "anthropic") {
			analysis = append(analysis, fmt.Sprintf("API endpoint modification: %s", entry))
		}
	}
	result["analysis"] = analysis

	return result, nil
}

// autostartProbe analyzes all auto-start mechanisms discovered
func (m Mediator) autostartProbe(report model.Report, input map[string]any) (any, error) {
	result := map[string]any{
		"services":   []map[string]any{},
		"crontab":    report.Discovery.CrontabLines,
		"registry":   []map[string]any{},
		"launchd":    []map[string]any{},
		"systemd":    []map[string]any{},
		"analysis":   []string{},
	}

	// Analyze services by platform
	for _, svc := range report.Discovery.Services {
		svcInfo := map[string]any{
			"name":    svc.Name,
			"scope":   svc.Scope,
			"path":    svc.Path,
		}
		switch svc.Platform {
		case "launchd":
			result["launchd"] = append(result["launchd"].([]map[string]any), svcInfo)
		case "systemd":
			result["systemd"] = append(result["systemd"].([]map[string]any), svcInfo)
		default:
			result["services"] = append(result["services"].([]map[string]any), svcInfo)
		}
	}

	// Add Windows registry auto-start entries
	if report.Host.OS == "windows" {
		for _, reg := range report.Discovery.RegistryKeys {
			if strings.Contains(reg.Path, "Run") || strings.Contains(reg.Path, "RunOnce") {
				result["registry"] = append(result["registry"].([]map[string]any), map[string]any{
					"root": reg.RootKey,
					"path": reg.Path,
					"data": reg.Data,
				})
			}
		}
	}

	// Generate analysis
	var analysis []string
	if len(result["launchd"].([]map[string]any)) > 0 {
		analysis = append(analysis, "macOS launchd services detected - check for agent auto-start")
	}
	if len(result["systemd"].([]map[string]any)) > 0 {
		analysis = append(analysis, "Linux systemd services detected - check for agent auto-start")
	}
	if len(result["registry"].([]map[string]any)) > 0 {
		analysis = append(analysis, "Windows registry auto-start entries detected")
	}
	if len(report.Discovery.CrontabLines) > 0 {
		analysis = append(analysis, "Crontab entries detected - check for scheduled agent tasks")
	}
	result["analysis"] = analysis

	return result, nil
}

func sliceResult[T any](key string, items []T, limit int) map[string]any {
	return map[string]any{
		key:     truncateSlice(items, limit),
		"count": len(items),
	}
}

func truncateSlice[T any](items []T, limit int) []T {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

func stringInput(input map[string]any, key string) (string, error) {
	raw, ok := input[key]
	if !ok {
		return "", fmt.Errorf("missing input %q", key)
	}
	value, ok := raw.(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("invalid input %q", key)
	}
	return value, nil
}

func allowedPathTarget(report model.Report, target string) bool {
	for _, group := range [][]string{
		report.Discovery.StateDirs,
		report.Discovery.WorkspaceDirs,
		report.Discovery.TempPaths,
		report.Discovery.AppPaths,
		report.Discovery.CLIPaths,
	} {
		if contains(group, target) {
			return true
		}
	}
	return false
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func runReadOnly(runner system.Runner, cmd []string) map[string]any {
	if len(cmd) == 0 {
		return map[string]any{"ok": false, "error": "missing command"}
	}
	result := runner.Run(context.Background(), cmd[0], cmd[1:]...)
	return map[string]any{
		"command": cmd,
		"ok":      result.OK,
		"code":    result.Code,
		"stdout":  truncateString(result.Stdout, 4000),
		"stderr":  truncateString(result.Stderr, 2000),
	}
}

func truncateString(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

func packageReadOnlyCommand(pkg model.PackageRef) []string {
	switch pkg.Manager {
	case "npm":
		return []string{"npm", "list", "-g", pkg.Name, "--depth=0"}
	case "pnpm":
		return []string{"pnpm", "list", "-g", pkg.Name, "--depth=0"}
	case "bun":
		return []string{"bun", "pm", "ls", "-g"}
	case "brew":
		args := []string{"brew", "info"}
		if pkg.Kind == "cask" {
			args = append(args, "--cask")
		}
		args = append(args, pkg.Name)
		return args
	default:
		return nil
	}
}

func toolSummary(report model.Report) map[string]any {
	return map[string]any{
		"product": report.Product,
		"command": report.Command,
		"host":    report.Host,
		"counts": map[string]int{
			"stateDirs":     len(report.Discovery.StateDirs),
			"workspaceDirs": len(report.Discovery.WorkspaceDirs),
			"services":      len(report.Discovery.Services),
			"packages":      len(report.Discovery.Packages),
			"processes":     len(report.Discovery.Processes),
			"containers":    len(report.Discovery.Containers),
			"images":        len(report.Discovery.Images),
			"planActions":   len(report.Plan.Actions),
			"evidenceItems": len(report.Evidence.Items),
			"confirmed":     len(report.Verify.Confirmed),
			"investigate":   len(report.Verify.Investigate),
		},
	}
}

func currentUID() string {
	if runtime.GOOS == "windows" {
		return "0"
	}
	if uid := strings.TrimSpace(os.Getenv("UID")); uid != "" {
		return uid
	}
	return "0"
}
