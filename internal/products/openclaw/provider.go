package openclaw

import (
	"context"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/skills"
	"github.com/tianrking/ClawRemove/internal/tools"
	"github.com/tianrking/ClawRemove/internal/verify"
)

type Provider struct{}

func (Provider) ID() string {
	return "openclaw"
}

func (Provider) DisplayName() string {
	return "OpenClaw"
}

func (Provider) Facts() model.ProductFacts {
	return model.ProductFacts{
		ID:          "openclaw",
		DisplayName: "OpenClaw",
		StateDirNames: []string{
			".openclaw",
			".clawdbot",
			".moldbot",
			".moltbot",
			".openclaw-legacy",
			".claw-dev",
		},
		ConfigNames: []string{
			"openclaw.json",
			"clawdbot.json",
			"moldbot.json",
			"moltbot.json",
		},
		Markers: []string{
			"openclaw",
			"ai.openclaw",
			"com.openclaw",
			"bot.molt",
			"clawdbot",
			"moltbot",
			"openclaw-legacy",
			"openclaw-beta",
		},
		ShellProfileGlobs: []string{
			".zshrc",
			".bashrc",
			".bash_profile",
			".profile",
			".config/fish/config.fish",
			".config/powershell/Microsoft.PowerShell_profile.ps1",
			"Documents/PowerShell/Microsoft.PowerShell_profile.ps1",
		},
		TempPrefixes: []string{
			"openclaw",
			"openclaw-",
			"openclaw-img-",
			"openclaw-updater-",
			"openclaw-restart-",
			"openclaw-zai-fallback-",
			".openclaw-install-stage-",
			".openclaw-install-backups",
		},
		AppPaths: []string{
			"/Applications/OpenClaw.app",
			"Applications/OpenClaw.app",
			"Library/Application Support/OpenClaw",
			"Library/Caches/OpenClaw",
			"Library/Preferences/ai.openclaw.mac.plist",
			"Library/Saved Application State/ai.openclaw.mac.savedState",
			"Library/WebKit/ai.openclaw.mac",
			"Library/HTTPStorages/ai.openclaw.mac",
			"Library/Cookies/ai.openclaw.mac.binarycookies",
			"Library/LaunchAgents/ai.openclaw.mac.plist",
			".local/share/applications/openclaw.desktop",
			".config/OpenClaw",
			"AppData/Roaming/OpenClaw",
			"AppData/Local/Programs/OpenClaw",
		},
		CLIPaths: []string{
			".local/bin/openclaw",
			".local/bin/openclaw.cmd",
		},
		PackageRefs: []model.PackageRef{
			{Manager: "npm", Name: "openclaw"},
			{Manager: "pnpm", Name: "openclaw"},
			{Manager: "bun", Name: "openclaw"},
			{Manager: "brew", Name: "openclaw-cli", Kind: "formula"},
			{Manager: "brew", Name: "openclaw", Kind: "formula"},
			{Manager: "brew", Name: "openclaw", Kind: "cask"},
		},
	}
}

func (p Provider) Capabilities() model.ProviderCapabilities {
	var caps model.ProviderCapabilities
	for _, skill := range p.Skills() {
		caps.Skills = append(caps.Skills, skill.Info())
	}
	for _, tool := range p.Tools() {
		caps.Tools = append(caps.Tools, tool.Info())
	}
	return caps
}

func (Provider) Skills() []skills.Skill {
	return []skills.Skill{
		residueAnalysisSkill{},
		safeRemovalReviewSkill{},
	}
}

func (Provider) VerificationRules() []verify.Rule {
	return []verify.Rule{
		shellProfileVerificationRule{},
	}
}

func (Provider) Tools() []tools.Tool {
	return []tools.Tool{
		stateProbeTool{},
		runtimeProbeTool{},
		shellProbeTool{},
	}
}

type shellProfileVerificationRule struct{}

func (shellProfileVerificationRule) Evaluate(residual *model.Residual) {
	if residual.Kind == "shell_profile" {
		// OpenClaw modifies shell profiles distinctly with its own exact CLI paths and variables.
		// If we matched the profile due to an exact marker (like .openclaw), we can trust it is strong evidence.
		residual.Evidence = "strong"
		residual.Rule = "verified-shell-profile"
		residual.Confidence = 0.85
	}
}

type residueAnalysisSkill struct{}

func (residueAnalysisSkill) Info() model.ProviderSkill {
	return model.ProviderSkill{
		ID:          "openclaw-residue-analysis",
		Name:        "OpenClaw Residue Analysis",
		Description: "Analyze OpenClaw residue using provider-specific state paths, package names, service markers, and legacy aliases.",
		Inputs:      []string{"discovery", "verification"},
	}
}

func (residueAnalysisSkill) Analyze(ctx context.Context, report model.Report) (any, error) {
	// For now, this just acts as a structural contract for the LLM.
	return nil, nil
}

type safeRemovalReviewSkill struct{}

func (safeRemovalReviewSkill) Info() model.ProviderSkill {
	return model.ProviderSkill{
		ID:          "openclaw-safe-removal-review",
		Name:        "OpenClaw Safe Removal Review",
		Description: "Review confirmed residue, high-risk actions, and investigate-only findings before apply.",
		Inputs:      []string{"verification", "plan", "advice"},
	}
}

func (safeRemovalReviewSkill) Analyze(ctx context.Context, report model.Report) (any, error) {
	return nil, nil
}

type stateProbeTool struct{}

func (stateProbeTool) Info() model.ProviderTool {
	return model.ProviderTool{
		ID:          "openclaw-state-probe",
		Name:        "OpenClaw State Probe",
		Description: "Read-only inspection of discovered state, workspace, temp, app, and CLI paths.",
		ReadOnly:    true,
		Targets:     []string{"state_dirs", "workspace_dirs", "path_probe"},
	}
}

func (stateProbeTool) Execute(ctx context.Context, report model.Report, input map[string]any) (any, error) {
	// This delegates execution to mediator fallback for generic probes inside ClawRemove core right now.
	// Over time, product-specific specific execution could be handled here directly, parsing `input`.
	return nil, nil // return nil so the generic Mediator can pick it up.
}

type runtimeProbeTool struct{}

func (runtimeProbeTool) Info() model.ProviderTool {
	return model.ProviderTool{
		ID:          "openclaw-runtime-probe",
		Name:        "OpenClaw Runtime Probe",
		Description: "Read-only inspection of discovered packages, services, listeners, and processes.",
		ReadOnly:    true,
		Targets:     []string{"services", "service_probe", "packages", "package_probe", "processes", "process_probe", "verification"},
	}
}

func (runtimeProbeTool) Execute(ctx context.Context, report model.Report, input map[string]any) (any, error) {
	return nil, nil
}

type shellProbeTool struct{}

func (shellProbeTool) Info() model.ProviderTool {
	return model.ProviderTool{
		ID:          "openclaw-shell-probe",
		Name:        "OpenClaw Shell Probe",
		Description: "Read-only inspection of shell profile traces and completion residue tied to OpenClaw markers.",
		ReadOnly:    true,
		Targets:     []string{"shell_profile_probe"},
	}
}

func (shellProbeTool) Execute(ctx context.Context, report model.Report, input map[string]any) (any, error) {
	return nil, nil
}
