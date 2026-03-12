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

// Runner is an interface for executing system commands.
// This allows for mocking in tests.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) CommandResult
	Exists(ctx context.Context, name string) bool
}

// RealRunner is the production implementation of Runner.
type RealRunner struct{}

// NewRunner creates a new Runner instance.
func NewRunner() Runner {
	return RealRunner{}
}

func (RealRunner) Run(ctx context.Context, name string, args ...string) CommandResult {
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

func (r RealRunner) Exists(ctx context.Context, name string) bool {
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
