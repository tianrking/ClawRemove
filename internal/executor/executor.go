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
	next := removeCompletionBlock(string(raw))
	if next == string(raw) {
		return model.Result{OK: true, Action: string(action.Kind), Target: action.Target, Reason: action.Reason, Skipped: true}
	}
	if options.DryRun {
		return model.Result{OK: true, Action: string(action.Kind), Target: action.Target, Reason: action.Reason, DryRun: true}
	}
	if err := os.WriteFile(action.Target, []byte(next), 0o644); err != nil {
		return model.Result{OK: false, Action: string(action.Kind), Target: action.Target, Reason: action.Reason, Error: err.Error()}
	}
	return model.Result{OK: true, Action: string(action.Kind), Target: action.Target, Reason: action.Reason}
}

func removeCompletionBlock(content string) string {
	lines := strings.Split(content, "\n")
	var out []string
	skipNext := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if skipNext {
			skipNext = false
			continue
		}
		if trimmed == "# OpenClaw Completion" {
			skipNext = true
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
