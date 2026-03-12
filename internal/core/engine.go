package core

import (
	"context"
	"fmt"
	"strings"

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

// InspectEnvironment generates a full environment report.
// This is the Agent Environment Inspector functionality.
func (e Engine) InspectEnvironment(ctx context.Context) model.EnvironmentReport {
	report := model.EnvironmentReport{
		Platform: e.host.OS,
		Host: model.Host{
			OS:      e.host.OS,
			Arch:    e.host.Arch,
			ExeExt:  e.host.ExeExt,
			HomeEnv: e.host.HomeEnv,
		},
	}

	// Run inventory scan
	invScanner := inventory.NewScanner(e.runner, e.host)
	inv := invScanner.Scan(ctx)

	// Convert inventory to report sections
	report.Runtime = e.buildRuntimeSection(inv)
	report.Agents = e.buildAgentsSection(inv)
	report.Artifacts = e.buildArtifactsSection(inv)

	// Run security scan
	secScanner := security.NewScanner()
	secReport := secScanner.Scan()
	report.Security = e.buildSecuritySection(secReport)

	// Build hygiene section
	report.Hygiene = e.buildHygieneSection(inv, report.Artifacts)

	return report
}

func (e Engine) buildRuntimeSection(inv inventory.AIInventory) model.RuntimeSection {
	var items []model.RuntimeItem
	for _, rt := range inv.Runtimes {
		var port int
		if len(rt.Ports) > 0 {
			port = rt.Ports[0]
		}
		items = append(items, model.RuntimeItem{
			Name:    rt.Name,
			Version: rt.Version,
			Path:    rt.Path,
			Running: rt.Running,
			Port:    port,
		})
	}

	summary := "No AI runtimes detected"
	if len(items) > 0 {
		running := 0
		for _, item := range items {
			if item.Running {
				running++
			}
		}
		summary = fmt.Sprintf("%d runtime(s) found, %d running", len(items), running)
	}

	return model.RuntimeSection{
		Detected: items,
		Summary:  summary,
	}
}

func (e Engine) buildAgentsSection(inv inventory.AIInventory) model.AgentsSection {
	var apps []model.AgentItem
	for _, agent := range inv.Agents {
		apps = append(apps, model.AgentItem{
			Name:    agent.Name,
			Path:    agent.Path,
			Version: agent.Version,
		})
	}

	var frameworks []model.AgentItem
	for _, fw := range inv.Frameworks {
		frameworks = append(frameworks, model.AgentItem{
			Name:    fw.Name,
			Version: fw.Version,
			Path:    fw.Path,
			Manager: fw.Manager,
		})
	}

	summary := "No AI agents detected"
	if len(apps) > 0 || len(frameworks) > 0 {
		parts := []string{}
		if len(apps) > 0 {
			parts = append(parts, fmt.Sprintf("%d application(s)", len(apps)))
		}
		if len(frameworks) > 0 {
			parts = append(parts, fmt.Sprintf("%d framework(s)", len(frameworks)))
		}
		summary = strings.Join(parts, ", ") + " found"
	}

	return model.AgentsSection{
		Applications: apps,
		Frameworks:   frameworks,
		Summary:      summary,
	}
}

func (e Engine) buildArtifactsSection(inv inventory.AIInventory) model.ArtifactsSection {
	var models []model.ArtifactItem
	var caches []model.ArtifactItem
	var vectorDBs []model.ArtifactItem
	var totalSize int64

	for _, cache := range inv.ModelCaches {
		item := model.ArtifactItem{
			Name: cache.Name,
			Path: cache.Path,
			Size: cache.Size,
		}
		if strings.Contains(strings.ToLower(cache.Name), "model") {
			models = append(models, item)
		} else {
			caches = append(caches, item)
		}
		totalSize += cache.Size
	}

	for _, vs := range inv.VectorStores {
		vectorDBs = append(vectorDBs, model.ArtifactItem{
			Name: vs.Name,
			Path: vs.Path,
		})
	}

	summary := "No AI artifacts detected"
	if totalSize > 0 {
		summary = fmt.Sprintf("Total: %s", formatSize(totalSize))
	}

	return model.ArtifactsSection{
		Models:    models,
		VectorDBs: vectorDBs,
		Caches:    caches,
		TotalSize: totalSize,
		Summary:   summary,
	}
}

func (e Engine) buildSecuritySection(secReport security.SecurityReport) model.SecuritySection {
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

func (e Engine) buildHygieneSection(inv inventory.AIInventory, artifacts model.ArtifactsSection) model.HygieneSection {
	var modelsSize, cacheSize, vectorDBSize, logSize int64
	var recommendations []string

	// Calculate sizes from model caches
	for _, cache := range inv.ModelCaches {
		name := strings.ToLower(cache.Name)
		if strings.Contains(name, "model") || strings.Contains(name, "ollama") || strings.Contains(name, "lmstudio") {
			modelsSize += cache.Size
		} else if strings.Contains(name, "cache") {
			cacheSize += cache.Size
		}
	}

	// Add recommendations based on findings
	if modelsSize > 10*1024*1024*1024 { // > 10GB
		recommendations = append(recommendations, "Consider cleaning old model versions to free up space")
	}
	if cacheSize > 5*1024*1024*1024 { // > 5GB
		recommendations = append(recommendations, "Large cache detected - review and clean unused caches")
	}

	totalSize := modelsSize + cacheSize + vectorDBSize + logSize

	summary := "Storage healthy"
	if totalSize > 50*1024*1024*1024 { // > 50GB
		summary = fmt.Sprintf("High storage usage: %s", formatSize(totalSize))
	} else if totalSize > 10*1024*1024*1024 { // > 10GB
		summary = fmt.Sprintf("Moderate storage usage: %s", formatSize(totalSize))
	}

	return model.HygieneSection{
		ModelsSize:     modelsSize,
		CacheSize:      cacheSize,
		VectorDBSize:   vectorDBSize,
		LogSize:        logSize,
		TotalSize:      totalSize,
		Recommendations: recommendations,
		Summary:        summary,
	}
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1fTB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
