package core

import (
	"context"

	"github.com/tianrking/ClawRemove/internal/discovery"
	"github.com/tianrking/ClawRemove/internal/executor"
	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/plan"
	"github.com/tianrking/ClawRemove/internal/products"
	"github.com/tianrking/ClawRemove/internal/system"
)

type Engine struct {
	runner system.Runner
}

func NewEngine(runner system.Runner) Engine {
	return Engine{runner: runner}
}

func (e Engine) Run(ctx context.Context, options model.Options) (model.Report, error) {
	provider, err := products.Resolve(options.Product)
	if err != nil {
		return model.Report{}, err
	}

	discoverer := discovery.New(e.runner, provider.Facts())
	discovered, err := discoverer.Discover(ctx)
	if err != nil {
		return model.Report{}, err
	}

	executionPlan := plan.Build(discovered, provider.Facts(), options)
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

	return model.Report{
		OK:        ok,
		Product:   provider.ID(),
		Command:   options.Command,
		DryRun:    options.DryRun,
		AuditOnly: options.AuditOnly,
		Discovery: discovered,
		Plan:      executionPlan,
		Results:   results,
	}, nil
}
