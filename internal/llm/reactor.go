package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/system"
)

type controlledAdvisor struct {
	client chatClient
	config Config
	runner system.Runner
}

type reactorStep struct {
	Kind            string                 `json:"kind"`
	ThoughtSummary  string                 `json:"thoughtSummary,omitempty"`
	Tool            string                 `json:"tool,omitempty"`
	Input           map[string]any         `json:"input,omitempty"`
	NeededEvidence  []string               `json:"neededEvidence,omitempty"`
	Recommendations []model.Recommendation `json:"recommendations,omitempty"`
	RiskNotes       []string               `json:"riskNotes,omitempty"`
	UserMessage     string                 `json:"userMessage,omitempty"`
}

func NewAdvisorFromEnv(runner system.Runner) Advisor {
	cfg := LoadConfigFromEnv()
	if !cfg.Enabled {
		return NewNoopAdvisor()
	}
	return controlledAdvisor{
		client: newClientFromConfig(cfg),
		config: cfg,
		runner: runner,
	}
}

func (a controlledAdvisor) Assess(ctx context.Context, report model.Report) model.Advice {
	base := NewNoopAdvisor().Assess(ctx, report)
	base.Mode = "react-controlled"
	base.UserMessage = "Controlled advisor is enabled. Review its recommendations as guidance, not as execution authority."

	systemPrompt := controlledSystemPrompt()
	messages := []chatMessage{
		{
			Role: "user",
			Content: fmt.Sprintf(
				"Analyze this ClawRemove report. You may request only read-only tools over the in-memory report.\nReport JSON:\n%s",
				mustJSON(reportContext(report)),
			),
		},
	}

	for step := 0; step < a.config.MaxSteps; step++ {
		content, err := a.client.CompleteJSON(ctx, systemPrompt, messages)
		if err != nil {
			base.RiskNotes = append(base.RiskNotes, "LLM advisor fallback: "+err.Error())
			return base
		}

		var next reactorStep
		if err := json.Unmarshal([]byte(content), &next); err != nil {
			base.RiskNotes = append(base.RiskNotes, "LLM advisor returned invalid JSON; deterministic fallback was used.")
			return base
		}

		switch strings.ToLower(next.Kind) {
		case "tool":
			toolResult, toolErr := a.executeTool(report, next.Tool, next.Input)
			if toolErr != nil {
				base.RiskNotes = append(base.RiskNotes, "LLM requested invalid tool input; deterministic fallback was used.")
				return base
			}
			messages = append(messages,
				chatMessage{Role: "assistant", Content: content},
				chatMessage{Role: "user", Content: fmt.Sprintf("Tool result for %s:\n%s", next.Tool, mustJSON(toolResult))},
			)
		case "final":
			return mergeAdvice(base, next)
		default:
			base.RiskNotes = append(base.RiskNotes, "LLM advisor returned an unsupported response kind; deterministic fallback was used.")
			return base
		}
	}

	base.RiskNotes = append(base.RiskNotes, "LLM advisor reached the maximum number of controlled reasoning steps.")
	return base
}

func controlledSystemPrompt() string {
	return strings.Join([]string{
		"You are ClawRemove Analyst, a controlled advisory model for a claw removal engine.",
		"You are not allowed to approve or execute destructive actions.",
		"You may only request read-only tools over the provided in-memory report.",
		"You must respond with JSON only.",
		"Valid response forms:",
		`{"kind":"tool","thoughtSummary":"...","tool":"summary|verification|state_dirs|workspace_dirs|services|packages|processes|containers|plan_actions|path_probe|service_probe|package_probe|process_probe|shell_profile_probe","input":{"limit":20}}`,
		`{"kind":"final","thoughtSummary":"...","neededEvidence":["..."],"riskNotes":["..."],"userMessage":"...","recommendations":[{"kind":"...","target":"...","reason":"...","risk":"low|medium|high","optIn":true,"evidence":"exact|strong|heuristic"}]}`,
		"Prefer final answers once you have enough evidence.",
		"If evidence is weak, say so and recommend review rather than deletion.",
		"You may only probe targets that already exist in the report.",
	}, "\n")
}

func reportContext(report model.Report) map[string]any {
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
			"confirmed":     len(report.Verify.Confirmed),
			"investigate":   len(report.Verify.Investigate),
		},
	}
}

func mergeAdvice(base model.Advice, next reactorStep) model.Advice {
	if next.ThoughtSummary != "" {
		base.ThoughtSummary = next.ThoughtSummary
	}
	if next.UserMessage != "" {
		base.UserMessage = next.UserMessage
	}
	if len(next.NeededEvidence) > 0 {
		base.NeededEvidence = next.NeededEvidence
	}
	if len(next.RiskNotes) > 0 {
		base.RiskNotes = append(base.RiskNotes, next.RiskNotes...)
	}
	if len(next.Recommendations) > 0 {
		base.Recommendations = next.Recommendations
	}
	return base
}

func (a controlledAdvisor) executeTool(report model.Report, tool string, input map[string]any) (any, error) {
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

	switch tool {
	case "summary":
		return reportContext(report), nil
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
	case "path_probe":
		return a.pathProbe(report, input)
	case "shell_profile_probe":
		return a.shellProfileProbe(report, input)
	case "service_probe":
		return a.serviceProbe(report, input)
	case "package_probe":
		return a.packageProbe(report, input)
	case "process_probe":
		return a.processProbe(report, input)
	default:
		return nil, fmt.Errorf("unsupported tool: %s", tool)
	}
}

func (a controlledAdvisor) pathProbe(report model.Report, input map[string]any) (any, error) {
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

func (a controlledAdvisor) shellProfileProbe(report model.Report, input map[string]any) (any, error) {
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

func (a controlledAdvisor) serviceProbe(report model.Report, input map[string]any) (any, error) {
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
		switch svc.Platform {
		case "linux":
			args := []string{"status", svc.Name + ".service", "--no-pager"}
			if svc.Scope == "user" {
				args = append([]string{"--user"}, args...)
			}
			cmd := append([]string{"systemctl"}, args...)
			result["status"] = runReadOnly(a.runner, cmd)
		case "darwin":
			result["status"] = runReadOnly(a.runner, []string{"launchctl", "print", "gui/" + itoa(os.Getuid()) + "/" + svc.Name})
		case "windows":
			result["status"] = runReadOnly(a.runner, []string{"schtasks", "/Query", "/TN", svc.Name, "/V", "/FO", "LIST"})
		}
		return result, nil
	}
	return nil, fmt.Errorf("service target is not present in the report")
}

func (a controlledAdvisor) packageProbe(report model.Report, input map[string]any) (any, error) {
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
			result["query"] = runReadOnly(a.runner, cmd)
		}
		return result, nil
	}
	return nil, fmt.Errorf("package target is not present in the report")
}

func (a controlledAdvisor) processProbe(report model.Report, input map[string]any) (any, error) {
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
		if proc.PID > 0 {
			result["status"] = runReadOnly(a.runner, []string{"ps", "-p", itoa(proc.PID), "-o", "pid=,ppid=,etime=,command="})
		}
		return result, nil
	}
	return nil, fmt.Errorf("process target is not present in the report")
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

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var buf [32]byte
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + (value % 10))
		value /= 10
	}
	return string(buf[i:])
}

func mustJSON(value any) string {
	body, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(body)
}
