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

	// Show help if no command or help requested
	if options.Command == "help" {
		fmt.Fprintln(stdout, usage)
		return 0, nil
	}

	if options.Command == "products" {
		var facts []model.ProductFacts
		for _, p := range products.Registry() {
			facts = append(facts, model.ProductFacts{ID: p.ID(), DisplayName: p.DisplayName()})
		}
		return 0, output.PrintProducts(stdout, facts, options.JSON)
	}

	// Environment inspection commands (no product required)
	if options.Command == "environment" || options.Command == "inventory" || options.Command == "security" || options.Command == "hygiene" {
		runner := system.NewRunner()
		host := platform.Detect()
		engine := core.NewEngine(runner, llm.NewAdvisorFromEnv(runner, host, nil), host)
		envReport := engine.InspectEnvironment(ctx)

		switch options.Command {
		case "inventory":
			return 0, output.PrintInventory(stdout, envReport, options.JSON)
		case "security":
			return 0, output.PrintSecurity(stdout, envReport, options.JSON)
		case "hygiene":
			return 0, output.PrintHygiene(stdout, envReport, options.JSON)
		case "environment":
			return 0, output.PrintEnvironment(stdout, envReport, options.JSON)
		}
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

const usage = `claw-remove - Agent Environment Inspector

USAGE:
    claw-remove <command> [options]

COMMANDS:
    environment    Full environment inspection report
    inventory      AI runtime and agent inventory
    security       AI tool security audit
    hygiene        AI storage usage analysis
    products       List supported product providers
    audit          Discover residuals for a product
    plan           Generate deletion plan (dry-run)
    apply          Execute cleanup actions
    verify         Verify cleanup results
    explain        Explain findings with AI analysis

OPTIONS:
    --product <id>     Product provider (default: openclaw)
    --dry-run          Report actions without executing
    --yes              Skip interactive confirmation
    --json             JSON output
    --ai               Include AI advisory analysis
    --audit            Audit only, no removal
    --keep-cli         Keep CLI packages
    --keep-app         Keep app bundles
    --keep-workspace   Keep workspace directories
    --keep-shell       Keep shell integration
    --kill-processes   Terminate matching processes
    --remove-docker    Remove docker/podman artifacts
    --version          Print version
    --help, -h        Show this help

EXAMPLES:
    claw-remove environment
    claw-remove audit --product openclaw
    claw-remove plan --product openclaw
    claw-remove apply --product openclaw --dry-run
    claw-remove apply --product openclaw --yes

Run 'claw-remove <command> --help' for command details.`

func parseOptions(args []string) (model.Options, error) {
	var opts model.Options
	opts.Command = "help"
	opts.Product = "openclaw"

	// Handle help flags anywhere
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "help" {
			return opts, nil
		}
	}

	if len(args) > 0 {
		switch args[0] {
		case "products":
			opts.Command = "products"
			return opts, nil
		case "environment", "inventory", "security", "hygiene":
			opts.Command = args[0]
			args = args[1:] // Continue to parse flags
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

