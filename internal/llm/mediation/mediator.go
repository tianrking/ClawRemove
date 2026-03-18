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
	case "search_agent_traces":
		return m.searchAgentTraces(report, input)
	case "quick_scan":
		return m.quickScan(report, input)
	case "config_probe":
		return m.configProbe(report, input)
	case "credential_probe":
		return m.credentialProbe(report, input)
	case "file_content_search":
		return m.fileContentSearch(report, input)
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

// searchAgentTraces searches for agent-specific patterns across the filesystem
func (m Mediator) searchAgentTraces(report model.Report, input map[string]any) (any, error) {
	productMarkers := []string{report.Product}
	if len(productMarkers) == 0 {
		productMarkers = []string{"agent", "bot", "assistant"}
	}

	result := map[string]any{
		"product":       report.Product,
		"markers":       productMarkers,
		"traces":        []map[string]any{},
		"analysis":      []string{},
		"searchedPaths": []string{},
	}

	// Search paths that commonly contain agent traces
	searchPaths := []struct {
		path     string
		category string
	}{
		{"/etc/hosts", "network"},
		{"/etc/environment", "environment"},
		{"/etc/profile", "shell"},
		{"/etc/profile.d", "shell"},
	}

	// Add user-specific paths
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		userPaths := []struct {
			path     string
			category string
		}{
			{filepath.Join(homeDir, ".ssh", "config"), "ssh"},
			{filepath.Join(homeDir, ".ssh", "authorized_keys"), "ssh"},
			{filepath.Join(homeDir, ".gitconfig"), "git"},
			{filepath.Join(homeDir, ".npmrc"), "npm"},
			{filepath.Join(homeDir, ".pypirc"), "python"},
			{filepath.Join(homeDir, ".config"), "config"},
		}
		searchPaths = append(searchPaths, userPaths...)
	}

	var traces []map[string]any
	var analysis []string

	for _, sp := range searchPaths {
		result["searchedPaths"] = append(result["searchedPaths"].([]string), sp.path)

		info, err := os.Stat(sp.path)
		if err != nil {
			continue
		}

		if info.IsDir() {
			// Search directory for matching files
			filepath.Walk(sp.path, func(path string, fi os.FileInfo, err error) error {
				if err != nil || fi.IsDir() {
					return nil
				}
				if content, err := os.ReadFile(path); err == nil {
					contentStr := string(content)
					for _, marker := range productMarkers {
						if strings.Contains(strings.ToLower(contentStr), strings.ToLower(marker)) {
							traces = append(traces, map[string]any{
								"path":     path,
								"category": sp.category,
								"marker":   marker,
								"size":     fi.Size(),
							})
							analysis = append(analysis, fmt.Sprintf("Found '%s' in %s", marker, path))
							break
						}
					}
				}
				return nil
			})
		} else {
			// Search single file
			if content, err := os.ReadFile(sp.path); err == nil {
				contentStr := string(content)
				for _, marker := range productMarkers {
					if strings.Contains(strings.ToLower(contentStr), strings.ToLower(marker)) {
						traces = append(traces, map[string]any{
							"path":     sp.path,
							"category": sp.category,
							"marker":   marker,
							"size":     info.Size(),
						})
						analysis = append(analysis, fmt.Sprintf("Found '%s' in %s", marker, sp.path))
						break
					}
				}
			}
		}
	}

	// Limit results
	limit := 20
	if len(traces) > limit {
		traces = traces[:limit]
	}
	if len(analysis) > limit {
		analysis = analysis[:limit]
	}

	result["traces"] = traces
	result["analysis"] = analysis
	result["traceCount"] = len(traces)

	return result, nil
}

// quickScan performs a fast scan of common sensitive directories
func (m Mediator) quickScan(report model.Report, input map[string]any) (any, error) {
	result := map[string]any{
		"platform":       report.Host.OS,
		"sensitivePaths": []map[string]any{},
		"summary":        map[string]int{},
	}

	homeDir, _ := os.UserHomeDir()

	// Define sensitive paths to check
	sensitiveChecks := []struct {
		path        string
		description string
		risk        string
		check       func(string) map[string]any
	}{
		{
			path:        "/etc/hosts",
			description: "Hosts file - network DNS overrides",
			risk:        "high",
			check: func(p string) map[string]any {
				if content, err := os.ReadFile(p); err == nil {
					lines := strings.Split(string(content), "\n")
					var nonStandard []string
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if line != "" && !strings.HasPrefix(line, "#") &&
							!strings.HasPrefix(line, "127.0.0.1 localhost") &&
							!strings.HasPrefix(line, "::1 localhost") {
							nonStandard = append(nonStandard, line)
						}
					}
					return map[string]any{"nonStandardEntries": nonStandard, "count": len(nonStandard)}
				}
				return nil
			},
		},
		{
			path:        filepath.Join(homeDir, ".ssh"),
			description: "SSH configuration directory",
			risk:        "high",
			check: func(p string) map[string]any {
				if entries, err := os.ReadDir(p); err == nil {
					var files []string
					for _, e := range entries {
						files = append(files, e.Name())
					}
					return map[string]any{"files": files, "count": len(files)}
				}
				return nil
			},
		},
	}

	var foundPaths []map[string]any
	summary := map[string]int{"total": 0, "modified": 0, "highRisk": 0}

	for _, check := range sensitiveChecks {
		info, err := os.Stat(check.path)
		if err != nil {
			continue
		}

		summary["total"]++
		pathInfo := map[string]any{
			"path":        check.path,
			"description": check.description,
			"risk":        check.risk,
			"exists":      true,
			"isDir":       info.IsDir(),
			"size":        info.Size(),
		}

		if checkResult := check.check(check.path); checkResult != nil {
			pathInfo["details"] = checkResult
			summary["modified"]++
		}

		if check.risk == "high" {
			summary["highRisk"]++
		}

		foundPaths = append(foundPaths, pathInfo)
	}

	result["sensitivePaths"] = foundPaths
	result["summary"] = summary

	return result, nil
}

// configProbe analyzes configuration files for agent modifications
func (m Mediator) configProbe(report model.Report, input map[string]any) (any, error) {
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

	result := map[string]any{
		"configs":  []map[string]any{},
		"analysis": []string{},
	}

	homeDir, _ := os.UserHomeDir()
	productMarkers := []string{report.Product}

	// Common config file patterns
	configPatterns := []string{
		filepath.Join(homeDir, ".config", "*", "config.json"),
		filepath.Join(homeDir, ".config", "*", "settings.json"),
		filepath.Join(homeDir, "*", "config.json"),
		filepath.Join(homeDir, ".env"),
	}

	var configs []map[string]any
	var analysis []string

	for _, pattern := range configPatterns {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			if len(configs) >= limit {
				break
			}

			content, err := os.ReadFile(match)
			if err != nil {
				continue
			}

			configInfo := map[string]any{
				"path":    match,
				"size":    len(content),
				"markers": []string{},
			}

			contentStr := string(content)
			for _, marker := range productMarkers {
				if strings.Contains(strings.ToLower(contentStr), strings.ToLower(marker)) {
					configInfo["markers"] = append(configInfo["markers"].([]string), marker)
				}
			}

			// Check for API keys
			if strings.Contains(contentStr, "API_KEY") || strings.Contains(contentStr, "api_key") ||
				strings.Contains(contentStr, "SECRET") || strings.Contains(contentStr, "secret") {
				analysis = append(analysis, fmt.Sprintf("Potential secrets in: %s", match))
				configInfo["hasSecrets"] = true
			}

			if len(configInfo["markers"].([]string)) > 0 || configInfo["hasSecrets"] == true {
				configs = append(configs, configInfo)
			}
		}
	}

	result["configs"] = configs
	result["analysis"] = analysis
	result["count"] = len(configs)

	return result, nil
}

// credentialProbe detects exposed credentials and API keys
func (m Mediator) credentialProbe(report model.Report, input map[string]any) (any, error) {
	result := map[string]any{
		"findings":       []map[string]any{},
		"riskLevel":      "unknown",
		"recommendation": "",
	}

	homeDir, _ := os.UserHomeDir()

	// Known credential patterns
	credPatterns := []struct {
		name     string
		pattern  string
		fileGlob string
	}{
		{"OpenAI API Key", "sk-", filepath.Join(homeDir, ".*")},
		{"Anthropic API Key", "sk-ant-", filepath.Join(homeDir, ".*")},
		{"Generic API Key", "API_KEY=", filepath.Join(homeDir, "*")},
		{"Environment File", ".env", filepath.Join(homeDir, "**/.env")},
	}

	var findings []map[string]any
	highRisk := false

	// Check state directories for credentials
	for _, dir := range report.Discovery.StateDirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			contentStr := string(content)
			for _, cp := range credPatterns {
				if strings.Contains(contentStr, cp.pattern) {
					findings = append(findings, map[string]any{
						"type":        cp.name,
						"file":        path,
						"pattern":     cp.pattern,
						"risk":        "high",
						"description": "Potential credential exposure",
					})
					highRisk = true
				}
			}
			return nil
		})
	}

	// Check environment variables
	for _, env := range report.Discovery.EnvVars {
		upperName := strings.ToUpper(env.Name)
		if strings.Contains(upperName, "API_KEY") ||
			strings.Contains(upperName, "SECRET") ||
			strings.Contains(upperName, "TOKEN") ||
			strings.Contains(upperName, "PASSWORD") {
			findings = append(findings, map[string]any{
				"type":        "Environment Variable",
				"name":        env.Name,
				"risk":        "medium",
				"description": "Sensitive environment variable detected",
			})
		}
	}

	result["findings"] = findings
	if highRisk {
		result["riskLevel"] = "high"
		result["recommendation"] = "Immediately rotate exposed API keys and secrets"
	} else if len(findings) > 0 {
		result["riskLevel"] = "medium"
		result["recommendation"] = "Review and secure detected credentials"
	} else {
		result["riskLevel"] = "low"
		result["recommendation"] = "No exposed credentials detected"
	}

	return result, nil
}

// fileContentSearch searches for specific patterns in files
func (m Mediator) fileContentSearch(report model.Report, input map[string]any) (any, error) {
	pattern, err := stringInput(input, "pattern")
	if err != nil {
		return nil, err
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

	result := map[string]any{
		"pattern":  pattern,
		"matches":  []map[string]any{},
		"count":    0,
		"analysis": []string{},
	}

	var matches []map[string]any
	var analysis []string

	// Search in discovered paths
	searchPaths := [][]string{
		report.Discovery.StateDirs,
		report.Discovery.WorkspaceDirs,
		report.Discovery.TempPaths,
	}

	for _, pathGroup := range searchPaths {
		for _, basePath := range pathGroup {
			filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() || len(matches) >= limit {
					return nil
				}

				content, err := os.ReadFile(path)
				if err != nil {
					return nil
				}

				contentStr := string(content)
				if strings.Contains(strings.ToLower(contentStr), strings.ToLower(pattern)) {
					// Find context around match
					idx := strings.Index(strings.ToLower(contentStr), strings.ToLower(pattern))
					contextStart := max(0, idx-50)
					contextEnd := min(len(contentStr), idx+len(pattern)+50)
					context := contentStr[contextStart:contextEnd]

					matches = append(matches, map[string]any{
						"path":    path,
						"context": context,
						"size":    info.Size(),
					})

					analysis = append(analysis, fmt.Sprintf("Found '%s' in %s", pattern, path))
				}
				return nil
			})
		}
	}

	result["matches"] = matches
	result["count"] = len(matches)
	result["analysis"] = analysis

	return result, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
