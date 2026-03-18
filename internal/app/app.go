package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tianrking/ClawRemove/internal/backup"
	"github.com/tianrking/ClawRemove/internal/cleanup"
	"github.com/tianrking/ClawRemove/internal/core"
	"github.com/tianrking/ClawRemove/internal/executor"
	"github.com/tianrking/ClawRemove/internal/llm"
	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/output"
	"github.com/tianrking/ClawRemove/internal/platform"
	"github.com/tianrking/ClawRemove/internal/products"
	"github.com/tianrking/ClawRemove/internal/system"
	"github.com/tianrking/ClawRemove/internal/tui"
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
	// If no command provided but stdin is terminal, launch TUI
	if options.Command == "help" {
		if len(args) == 0 {
			options.Command = "tui"
		} else {
			fmt.Fprintln(stdout, usage)
			return 0, nil
		}
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

	// Cleanup command (no product required)
	if options.Command == "cleanup" {
		runner := system.NewRunner()
		scanner := cleanup.NewScanner(runner)
		report := scanner.ScanAll(ctx)

		// Filter by category if specified
		if options.Category != "" && options.Category != "all" {
			var filtered []model.CleanupCandidate
			for _, c := range report.Candidates {
				if c.Category == options.Category {
					filtered = append(filtered, c)
				}
			}
			report.Candidates = filtered
			var totalReclaimable int64
			for _, c := range filtered {
				totalReclaimable += c.Size
			}
			report.TotalReclaimable = totalReclaimable
			if len(filtered) == 0 {
				report.Summary = "No cleanup candidates found for category: " + options.Category
			}
		}

		return 0, output.PrintCleanup(stdout, report, options.JSON)
	}
	
	if options.Command == "tui" {
		runner := system.NewRunner()
		host := platform.Detect()
		engine := core.NewEngine(runner, llm.NewAdvisorFromEnv(runner, host, nil), host)
		scanner := cleanup.NewScanner(runner)
		homeDir, _ := os.UserHomeDir()
		backupMgr := backup.NewManager(filepath.Join(homeDir, ".clawremove"))
		exec := executor.New(runner, backupMgr)
		
		if err := tui.Start(engine, scanner, exec, options); err != nil {
			return 1, err
		}
		return 0, nil
	}

	// Backup commands
	if options.Command == "snapshots" || options.Command == "rollback" {
		homeDir, _ := os.UserHomeDir()
		mgr := backup.NewManager(filepath.Join(homeDir, ".clawremove"))

		if options.Command == "snapshots" {
			snaps, err := mgr.ListSnapshots()
			if err != nil {
				return 1, err
			}
			return 0, output.PrintSnapshots(stdout, snaps, options.JSON)
		}

		if options.Command == "rollback" {
			if options.SnapshotID == "" {
				return 1, errors.New("rollback requires --id flag with snapshot ID")
			}
			if err := mgr.Rollback(options.SnapshotID); err != nil {
				return 1, fmt.Errorf("rollback failed: %w", err)
			}
			if !options.JSON && !options.Quiet {
				fmt.Fprintf(stdout, "✅ Successfully rolled back to snapshot %s\n", options.SnapshotID)
			}
			return 0, nil
		}
	}

	// Handle --product all to run across all registered products
	if options.Product == "all" {
		return runAllProducts(ctx, options, stdout, stderr)
	}

	provider, err := products.Resolve(options.Product)
	if err != nil {
		return 2, err
	}

	runner := system.NewRunner()
	host := platform.Detect()
	engine := core.NewEngine(runner, llm.NewAdvisorFromEnv(runner, host, provider.Tools()), host)

	// Create streaming function - only stream to stdout if not in JSON mode
	streamFunc := core.NilStreamFunc
	if !options.JSON && !options.Quiet {
		streamFunc = func(format string, args ...any) {
			if format == "" {
				fmt.Fprintln(stdout)
			} else {
				fmt.Fprintf(stdout, format+"\n", args...)
			}
		}
	}

	if options.Command == "apply" && !options.DryRun && !options.Yes {
		if options.JSON {
			return 2, errors.New("interactive apply requires a TTY-style confirmation; rerun without --json or use --yes after reviewing plan")
		}
		preview := options
		preview.DryRun = true
		report, err := engine.RunWithStream(ctx, preview, streamFunc)
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
	report, err := engine.RunWithStream(ctx, options, streamFunc)
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

// runAllProducts executes the command across all registered products
func runAllProducts(ctx context.Context, options model.Options, stdout io.Writer, stderr io.Writer) (int, error) {
	registry := products.Registry()
	if len(registry) == 0 {
		return 2, errors.New("no products registered")
	}

	// Create streaming function
	streamFunc := core.NilStreamFunc
	if !options.JSON && !options.Quiet {
		streamFunc = func(format string, args ...any) {
			if format == "" {
				fmt.Fprintln(stdout)
			} else {
				fmt.Fprintf(stdout, format+"\n", args...)
			}
		}
	}

	var allReports []model.Report
	var hasErrors bool

	streamFunc("🔍 Scanning all %d registered products...", len(registry))
	streamFunc("")

	for i, p := range registry {
		productID := p.ID()
		streamFunc("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		streamFunc("[%d/%d] %s (%s)", i+1, len(registry), p.DisplayName(), productID)
		streamFunc("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Create options for this specific product
		productOpts := options
		productOpts.Product = productID

		runner := system.NewRunner()
		host := platform.Detect()
		engine := core.NewEngine(runner, llm.NewAdvisorFromEnv(runner, host, p.Tools()), host)

		report, err := engine.RunWithStream(ctx, productOpts, streamFunc)
		if err != nil {
			streamFunc("❌ Error: %s", err.Error())
			hasErrors = true
			continue
		}

		if !report.OK {
			hasErrors = true
		}

		if err := output.PrintReport(stdout, report, options.JSON); err != nil {
			return 1, err
		}

		allReports = append(allReports, report)
		streamFunc("")
	}

	// Print summary
	streamFunc("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	streamFunc("📊 Summary: Scanned %d products", len(registry))
	if hasErrors {
		streamFunc("⚠️  Some products had issues (see above)")
		return 1, errors.New("one or more products had issues")
	}
	streamFunc("✅ All products processed successfully")
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
    cleanup        Scan and clean old models, caches, vector DBs, and logs
    tui            Launch interactive Terminal User Interface
    products       List supported product providers
    snapshots      List available backup snapshots
    rollback       Restore a state from a backup snapshot
    audit          Discover residuals for a product
    plan           Generate deletion plan (dry-run)
    apply          Execute cleanup actions
    verify         Verify cleanup results
    explain        Explain findings with AI analysis

OPTIONS:
    --product <id>     Product provider (default: openclaw, use 'all' for all products)
    --category <cat>   Cleanup category filter (model_version/orphaned_cache/unused_vectordb/log_rotation/all)
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
    claw-remove cleanup
    claw-remove cleanup --category model_version
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
		case "environment", "inventory", "security", "hygiene", "cleanup", "tui":
			opts.Command = args[0]
			args = args[1:] // Continue to parse flags
		case "audit", "plan", "apply", "verify", "explain", "snapshots", "rollback":
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
	fs.BoolVar(&opts.NoBackup, "no-backup", false, "skip taking safety backup before applying removals")
	fs.StringVar(&opts.Category, "category", "", "cleanup category filter (model_version/orphaned_cache/unused_vectordb/log_rotation/all)")
	fs.StringVar(&opts.Product, "product", opts.Product, "product provider id")
	fs.StringVar(&opts.SnapshotID, "id", "", "snapshot id for rollback")
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

