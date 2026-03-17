package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tianrking/ClawRemove/internal/backup"
	"github.com/tianrking/ClawRemove/internal/discovery"
	"github.com/tianrking/ClawRemove/internal/evidence"
	"github.com/tianrking/ClawRemove/internal/executor"
	"github.com/tianrking/ClawRemove/internal/inventory"
	"github.com/tianrking/ClawRemove/internal/llm"
	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/plan"
	"github.com/tianrking/ClawRemove/internal/platform"
	"github.com/tianrking/ClawRemove/internal/products"
	"github.com/tianrking/ClawRemove/internal/security"
	"github.com/tianrking/ClawRemove/internal/system"
	"github.com/tianrking/ClawRemove/internal/verify"
)

type Engine struct {
	runner  system.Runner
	advisor llm.Advisor
	host    platform.Host
}

func NewEngine(runner system.Runner, advisor llm.Advisor, host platform.Host) Engine {
	return Engine{runner: runner, advisor: advisor, host: host}
}

// StreamFunc is an alias for the llm.StreamFunc type for convenience.
type StreamFunc = llm.StreamFunc

// NilStreamFunc is a no-op streaming function.
var NilStreamFunc = llm.NilStreamFunc

func (e Engine) Run(ctx context.Context, options model.Options) (model.Report, error) {
	return e.RunWithStream(ctx, options, NilStreamFunc)
}

func (e Engine) RunWithStream(ctx context.Context, options model.Options, stream StreamFunc) (model.Report, error) {
	provider, err := products.Resolve(options.Product)
	if err != nil {
		return model.Report{}, err
	}

	discoverer := discovery.New(e.runner, provider.Facts(), e.host)
	discovered, err := discoverer.Discover(ctx)
	if err != nil {
		return model.Report{}, err
	}

	evidenceSet := evidence.Build(discovered, provider.Facts())
	executionPlan := plan.Build(discovered, evidenceSet, provider.Facts(), options, e.host)
	verification := verify.Classify(evidenceSet, provider.VerificationRules())
	var results []model.Result
	if options.Command == "apply" && !options.AuditOnly {
		homeDir, _ := os.UserHomeDir()
		backupMgr := backup.NewManager(filepath.Join(homeDir, ".clawremove"))
		exec := executor.New(e.runner, backupMgr)
		results = exec.Execute(ctx, executionPlan, options, provider.ID())
	}

	ok := true
	for _, result := range results {
		if !result.OK {
			ok = false
			break
		}
	}
	if options.Command != "apply" {
		ok = true
	}

	var advice *model.Advice
	if options.AI || options.Command == "explain" {
		assessed := e.advisor.AssessWithStream(ctx, model.Report{
			OK:        ok,
			Product:   provider.ID(),
			Command:   options.Command,
			DryRun:    options.DryRun,
			AuditOnly: options.AuditOnly,
			Host: model.Host{
				OS:      e.host.OS,
				Arch:    e.host.Arch,
				ExeExt:  e.host.ExeExt,
				HomeEnv: e.host.HomeEnv,
			},
			Capabilities: provider.Capabilities(),
			Discovery:    discovered,
			Evidence:     evidenceSet,
			Verify:       verification,
			Plan:         executionPlan,
			Results:      results,
		}, provider.Skills(), stream)
		advice = &assessed
	}

	return model.Report{
		OK:        ok,
		Product:   provider.ID(),
		Command:   options.Command,
		DryRun:    options.DryRun,
		AuditOnly: options.AuditOnly,
		Host: model.Host{
			OS:      e.host.OS,
			Arch:    e.host.Arch,
			ExeExt:  e.host.ExeExt,
			HomeEnv: e.host.HomeEnv,
		},
		Capabilities: provider.Capabilities(),
		Discovery:    discovered,
		Evidence:     evidenceSet,
		Verify:       verification,
		Plan:         executionPlan,
		Results:      results,
		Advice:       advice,
	}, nil
}

// InspectEnvironment generates a full environment report.
// This is the Agent Environment Inspector functionality.
func (e Engine) InspectEnvironment(ctx context.Context) model.EnvironmentReport {
	// Run inventory scan
	invScanner := inventory.NewScanner(e.runner, e.host)
	inv := invScanner.Scan(ctx)

	// Run security scan
	secScanner := security.NewScanner()
	secReport := secScanner.Scan()

	// Build report sections using dedicated builders
	artifacts := inventory.BuildArtifactsSection(inv)

	return model.EnvironmentReport{
		Platform: e.host.OS,
		Host: model.Host{
			OS:      e.host.OS,
			Arch:    e.host.Arch,
			ExeExt:  e.host.ExeExt,
			HomeEnv: e.host.HomeEnv,
		},
		Runtime:   inventory.BuildRuntimeSection(inv),
		Agents:    inventory.BuildAgentsSection(inv),
		Artifacts: artifacts,
		Security:  buildSecuritySection(secReport),
		Hygiene:   inventory.BuildHygieneSection(inv, artifacts),
	}
}

func buildSecuritySection(secReport security.SecurityReport) model.SecuritySection {
	var findings []model.SecurityFinding
	for _, f := range secReport.Findings {
		findings = append(findings, model.SecurityFinding{
			Type:        f.Type,
			Provider:    f.Provider,
			Location:    f.Location,
			Line:        f.Line,
			Severity:    f.Severity,
			Remediation: f.Remediation,
		})
	}

	summary := "No security issues found"
	if secReport.Summary.Total > 0 {
		if secReport.Summary.HighRisk > 0 {
			summary = fmt.Sprintf("%d issue(s) found, %d high risk", secReport.Summary.Total, secReport.Summary.HighRisk)
		} else {
			summary = fmt.Sprintf("%d issue(s) found", secReport.Summary.Total)
		}
	}

	return model.SecuritySection{
		Findings: findings,
		HighRisk: secReport.Summary.HighRisk,
		Summary:  summary,
	}
}
