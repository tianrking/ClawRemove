package core

import (
	"context"

	"github.com/tianrking/ClawRemove/internal/discovery"
	"github.com/tianrking/ClawRemove/internal/evidence"
	"github.com/tianrking/ClawRemove/internal/executor"
	"github.com/tianrking/ClawRemove/internal/llm"
	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/plan"
	"github.com/tianrking/ClawRemove/internal/platform"
	"github.com/tianrking/ClawRemove/internal/products"
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

func (e Engine) Run(ctx context.Context, options model.Options) (model.Report, error) {
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
		exec := executor.New(e.runner)
		results = exec.Execute(ctx, executionPlan, options)
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
		assessed := e.advisor.Assess(ctx, model.Report{
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
		}, provider.Skills())
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
