package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/tianrking/ClawRemove/internal/core"
	"github.com/tianrking/ClawRemove/internal/llm"
	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/output"
	"github.com/tianrking/ClawRemove/internal/platform"
	"github.com/tianrking/ClawRemove/internal/products"
	"github.com/tianrking/ClawRemove/internal/system"
)

func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) (int, error) {
	options, err := parseOptions(args)
	if err != nil {
		return 2, err
	}
	if options.Command == "products" {
		return 0, printProducts(stdout, options.JSON)
	}

	runner := system.NewRunner()
	engine := core.NewEngine(runner, llm.NewAdvisorFromEnv(runner), platform.Detect())
	report, err := engine.Run(ctx, options)
	if err != nil {
		return 1, err
	}
	if err := output.PrintReport(stdout, report, options.JSON); err != nil {
		return 1, err
	}
	if !report.OK {
		return 1, errors.New("one or more actions failed")
	}
	return 0, nil
}

func parseOptions(args []string) (model.Options, error) {
	var opts model.Options
	opts.Command = "apply"
	opts.Product = "openclaw"
	if len(args) > 0 {
		switch args[0] {
		case "products":
			opts.Command = "products"
			return opts, nil
		case "audit", "plan", "apply", "verify", "explain":
			opts.Command = args[0]
			args = args[1:]
		}
	}

	fs := flag.NewFlagSet("claw-remove", flag.ContinueOnError)
	fs.BoolVar(&opts.DryRun, "dry-run", false, "report actions without executing them")
	fs.BoolVar(&opts.Yes, "yes", false, "reserved for non-interactive confirmation flows")
	fs.BoolVar(&opts.Quiet, "quiet", false, "reduce console output")
	fs.BoolVar(&opts.JSON, "json", false, "emit machine-readable JSON output")
	fs.BoolVar(&opts.AI, "ai", false, "include controlled advisory analysis in the report")
	fs.BoolVar(&opts.AuditOnly, "audit", false, "only audit residuals, do not execute removal actions")
	fs.BoolVar(&opts.KeepCLI, "keep-cli", false, "keep CLI packages and wrappers")
	fs.BoolVar(&opts.KeepApp, "keep-app", false, "keep app bundles and app data")
	fs.BoolVar(&opts.KeepWorkspace, "keep-workspace", false, "keep workspace directories")
	fs.BoolVar(&opts.KeepShell, "keep-shell", false, "keep shell completion/profile integration")
	fs.BoolVar(&opts.KillProcesses, "kill-processes", false, "terminate matching processes")
	fs.BoolVar(&opts.RemoveDocker, "remove-docker", false, "remove matching docker/podman containers and images")
	fs.StringVar(&opts.Product, "product", opts.Product, "product provider id")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return model.Options{}, fmt.Errorf("parse flags: %w", err)
	}
	switch opts.Command {
	case "audit":
		opts.AuditOnly = true
	case "plan":
		opts.DryRun = true
	case "verify":
		opts.AuditOnly = true
	case "explain":
		opts.AuditOnly = true
		opts.AI = true
	}
	return opts, nil
}

func printProducts(w io.Writer, jsonMode bool) error {
	if jsonMode {
		_, err := io.WriteString(w, "[")
		if err != nil {
			return err
		}
		for i, provider := range products.Registry() {
			if i > 0 {
				if _, err := io.WriteString(w, ","); err != nil {
					return err
				}
			}
			if _, err := io.WriteString(w, fmt.Sprintf(`{"id":"%s","displayName":"%s"}`, provider.ID(), provider.DisplayName())); err != nil {
				return err
			}
		}
		_, err = io.WriteString(w, "]\n")
		return err
	}
	for _, provider := range products.Registry() {
		if _, err := io.WriteString(w, provider.ID()+"\t"+provider.DisplayName()+"\n"); err != nil {
			return err
		}
	}
	return nil
}
