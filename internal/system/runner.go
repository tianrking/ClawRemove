package system

import (
	"bytes"
	"context"
	"os/exec"
)

type CommandResult struct {
	OK     bool
	Code   int
	Stdout string
	Stderr string
}

type Runner struct{}

func NewRunner() Runner {
	return Runner{}
}

func (Runner) Run(ctx context.Context, name string, args ...string) CommandResult {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := CommandResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if err == nil {
		result.OK = true
		return result
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.Code = exitErr.ExitCode()
		return result
	}
	result.Code = -1
	return result
}

func (r Runner) Exists(ctx context.Context, name string) bool {
	checker := "which"
	args := []string{name}
	if isWindows() {
		checker = "where"
	}
	return r.Run(ctx, checker, args...).OK
}

func isWindows() bool {
	return exec.Command("cmd", "/c", "ver").Run() == nil
}
