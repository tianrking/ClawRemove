package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tianrking/ClawRemove/internal/core"
	"github.com/tianrking/ClawRemove/internal/llm"
	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/output"
	"github.com/tianrking/ClawRemove/internal/platform"
	"github.com/tianrking/ClawRemove/internal/products"
	"github.com/tianrking/ClawRemove/internal/system"
)

var Version = "dev"

func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) (int, error) {
	options, err := parseOptions(args)
	if err != nil {
		return 2, err
	}
	if options.Version {
		fmt.Fprintln(stdout, "claw-remove version", Version)
		return 0, nil
	}
	if options.Command == "products" {
		var facts []model.ProductFacts
		for _, p := range products.Registry() {
			facts = append(facts, model.ProductFacts{ID: p.ID(), DisplayName: p.DisplayName()})
		}
		return 0, output.PrintProducts(stdout, facts, options.JSON)
	}

	provider, err := products.Resolve(options.Product)
	if err != nil {
		return 2, err
	}

	runner := system.NewRunner()
	host := platform.Detect()
	engine := core.NewEngine(runner, llm.NewAdvisorFromEnv(runner, host, provider.Tools()), host)
	if options.Command == "apply" && !options.DryRun && !options.Yes {
		if options.JSON {
			return 2, errors.New("interactive apply requires a TTY-style confirmation; rerun without --json or use --yes after reviewing plan")
		}
		preview := options
		preview.DryRun = true
		report, err := engine.Run(ctx, preview)
		if err != nil {
			return 1, err
		}
		if err := output.PrintReport(stdout, report, false); err != nil {
			return 1, err
		}
		confirmed, err := output.ConfirmApply(os.Stdin, stdout, stderr, options.Product, report)
		if err != nil {
			return 1, err
		}
		if !confirmed {
			return 1, errors.New("apply cancelled by user")
		}
	}
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
	fs.BoolVar(&opts.Version, "version", false, "print version information and exit")
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

