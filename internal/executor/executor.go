package executor

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/system"
)

type Executor struct {
	runner system.Runner
}

func New(runner system.Runner) Executor {
	return Executor{runner: runner}
}

func (e Executor) Execute(ctx context.Context, plan model.Plan, options model.Options) []model.Result {
	var results []model.Result
	for _, action := range plan.Actions {
		switch action.Kind {
		case model.ActionReportOnly:
			results = append(results, model.Result{
				OK:      true,
				Action:  string(action.Kind),
				Target:  action.Target,
				Reason:  action.Reason,
				Skipped: true,
			})
		case model.ActionRunCommand:
			results = append(results, e.runCommand(ctx, action, options))
		case model.ActionRemovePath:
			results = append(results, e.removePath(action, options))
		case model.ActionEditFile:
			results = append(results, e.cleanShellProfile(action, options))
		}
	}
	return results
}

func (e Executor) runCommand(ctx context.Context, action model.Action, options model.Options) model.Result {
	if len(action.Command) == 0 {
		return model.Result{OK: false, Action: string(action.Kind), Target: action.Target, Reason: action.Reason, Error: "missing command"}
	}
	if options.DryRun {
		return model.Result{OK: true, Action: string(action.Kind), Target: strings.Join(action.Command, " "), Reason: action.Reason, DryRun: true}
	}
	cmd := slicesClone(action.Command)
	if runtime.GOOS == "darwin" && len(cmd) >= 3 && cmd[0] == "launchctl" && cmd[1] == "bootout" && strings.Contains(cmd[2], "$UID") {
		cmd[2] = strings.ReplaceAll(cmd[2], "$UID", itoa(os.Getuid()))
	}
	result := e.runner.Run(ctx, cmd[0], cmd[1:]...)
	if !result.OK {
		return model.Result{OK: false, Action: string(action.Kind), Target: strings.Join(cmd, " "), Reason: action.Reason, Error: strings.TrimSpace(result.Stderr)}
	}
	return model.Result{OK: true, Action: string(action.Kind), Target: strings.Join(cmd, " "), Reason: action.Reason}
}

func (e Executor) removePath(action model.Action, options model.Options) model.Result {
	target := filepath.Clean(action.Target)
	if target == string(filepath.Separator) {
		return model.Result{OK: false, Action: string(action.Kind), Target: target, Reason: action.Reason, Error: "refusing to remove root path"}
	}
	if options.DryRun {
		return model.Result{OK: true, Action: string(action.Kind), Target: target, Reason: action.Reason, DryRun: true}
	}
	if err := os.RemoveAll(target); err != nil {
		return model.Result{OK: false, Action: string(action.Kind), Target: target, Reason: action.Reason, Error: err.Error()}
	}
	return model.Result{OK: true, Action: string(action.Kind), Target: target, Reason: action.Reason}
}

func (e Executor) cleanShellProfile(action model.Action, options model.Options) model.Result {
	raw, err := os.ReadFile(action.Target)
	if err != nil {
		if os.IsNotExist(err) {
			return model.Result{OK: true, Action: string(action.Kind), Target: action.Target, Reason: action.Reason, Skipped: true}
		}
		return model.Result{OK: false, Action: string(action.Kind), Target: action.Target, Reason: action.Reason, Error: err.Error()}
	}
	next := removeMarkerLines(string(raw), action.Markers)
	if next == string(raw) {
		return model.Result{OK: true, Action: string(action.Kind), Target: action.Target, Reason: action.Reason, Skipped: true}
	}
	if options.DryRun {
		return model.Result{OK: true, Action: string(action.Kind), Target: action.Target, Reason: action.Reason, DryRun: true}
	}
	// Safety: write backup before mutating
	backupPath := action.Target + ".clawremove.bak"
	if berr := os.WriteFile(backupPath, raw, 0o644); berr != nil {
		return model.Result{OK: false, Action: string(action.Kind), Target: action.Target, Reason: action.Reason, Error: "could not write backup: " + berr.Error()}
	}
	if err := os.WriteFile(action.Target, []byte(next), 0o644); err != nil {
		return model.Result{OK: false, Action: string(action.Kind), Target: action.Target, Reason: action.Reason, Error: err.Error()}
	}
	return model.Result{OK: true, Action: string(action.Kind), Target: action.Target, Reason: action.Reason}
}

// removeMarkerLines removes any line from a shell profile that contains a provider marker,
// along with the immediately following non-empty eval/source line if paired with it.
func removeMarkerLines(content string, markers []string) string {
	lines := strings.Split(content, "\n")
	var out []string
	skipNext := false
	for i, line := range lines {
		if skipNext {
			skipNext = false
			continue
		}
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		// Drop comment or eval lines that reference a known marker
		markerLine := false
		for _, m := range markers {
			if m != "" && strings.Contains(lower, strings.ToLower(m)) {
				markerLine = true
				break
			}
		}
		// Legacy hard-coded OpenClaw single-line completion blocks (backward compat)
		if trimmed == "# OpenClaw Completion" || trimmed == "# openclaw completion" {
			markerLine = true
		}
		if markerLine {
			// If the next line is an eval or source that relates, skip it too
			if i+1 < len(lines) {
				nextTrimmed := strings.TrimSpace(lines[i+1])
				nextLower := strings.ToLower(nextTrimmed)
				if strings.HasPrefix(nextLower, "eval") || strings.HasPrefix(nextLower, "source") || strings.HasPrefix(nextLower, ". ") {
					for _, m := range markers {
						if m != "" && strings.Contains(nextLower, strings.ToLower(m)) {
							skipNext = true
							break
						}
					}
				}
			}
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}


func slicesClone(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	return out
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
