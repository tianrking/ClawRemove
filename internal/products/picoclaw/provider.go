package picoclaw

import (
	"context"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/skills"
	"github.com/tianrking/ClawRemove/internal/tools"
	"github.com/tianrking/ClawRemove/internal/verify"
)

type Provider struct{}

func (Provider) ID() string {
	return "picoclaw"
}

func (Provider) DisplayName() string {
	return "PicoClaw"
}

func (Provider) Facts() model.ProductFacts {
	return model.ProductFacts{
		ID:          "picoclaw",
		DisplayName: "PicoClaw",
		StateDirNames: []string{
			".picoclaw",
		},
		WorkspaceDirNames: []string{
			"workspace",
			"skills",
			"memory",
		},
		ConfigNames: []string{
			"config.json",
		},
		Markers: []string{
			"picoclaw",
			"pico.claw",
			"picoclaw-launcher",
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
			"picoclaw",
			"picoclaw-",
		},
		AppPaths: []string{
			".local/share/picoclaw",
			".config/picoclaw",
			"Library/Application Support/PicoClaw",
			"Library/Caches/PicoClaw",
			"AppData/Roaming/PicoClaw",
			"AppData/Local/PicoClaw",
			"AppData/Local/Programs/PicoClaw",
		},
		CLIPaths: []string{
			".local/bin/picoclaw",
			".local/bin/picoclaw-launcher",
		},
		PackageRefs: []model.PackageRef{
			// PicoClaw is distributed as binary releases
		},
		// PicoClaw's known gateway and IPC ports
		ListenerPorts: []int{18790, 18791},
		// Windows registry paths
		RegistryPaths: []string{
			"HKCU\\Software\\PicoClaw",
			"HKLM\\SOFTWARE\\PicoClaw",
			"HKLM\\SOFTWARE\\WOW6432Node\\PicoClaw",
		},
		// Environment variables
		EnvVarNames: []string{
			"PICOCLAW_HOME",
			"PICOCLAW_CONFIG",
			"PICOCLAW_BUILTIN_SKILLS",
			"PICOCLAW_GATEWAY_HOST",
			"PICOCLAW_AGENTS_DEFAULTS_MODEL",
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
	return []verify.Rule{}
}

func (Provider) Tools() []tools.Tool {
	return []tools.Tool{
		stateProbeTool{},
		runtimeProbeTool{},
	}
}

type residueAnalysisSkill struct{}

func (residueAnalysisSkill) Info() model.ProviderSkill {
	return model.ProviderSkill{
		ID:          "picoclaw-residue-analysis",
		Name:        "PicoClaw Residue Analysis",
		Description: "Analyze PicoClaw residue using provider-specific state paths and service markers.",
		Inputs:      []string{"discovery", "verification"},
	}
}

func (residueAnalysisSkill) Analyze(ctx context.Context, report model.Report) (any, error) {
	return nil, nil
}

type safeRemovalReviewSkill struct{}

func (safeRemovalReviewSkill) Info() model.ProviderSkill {
	return model.ProviderSkill{
		ID:          "picoclaw-safe-removal-review",
		Name:        "PicoClaw Safe Removal Review",
		Description: "Review confirmed residue and high-risk actions before apply.",
		Inputs:      []string{"verification", "plan", "advice"},
	}
}

func (safeRemovalReviewSkill) Analyze(ctx context.Context, report model.Report) (any, error) {
	return nil, nil
}

type stateProbeTool struct{}

func (stateProbeTool) Info() model.ProviderTool {
	return model.ProviderTool{
		ID:          "picoclaw-state-probe",
		Name:        "PicoClaw State Probe",
		Description: "Read-only inspection of discovered state and workspace paths.",
		ReadOnly:    true,
		Targets:     []string{"state_dirs", "workspace_dirs", "path_probe"},
	}
}

func (stateProbeTool) Execute(ctx context.Context, report model.Report, input map[string]any) (any, error) {
	return nil, nil
}

type runtimeProbeTool struct{}

func (runtimeProbeTool) Info() model.ProviderTool {
	return model.ProviderTool{
		ID:          "picoclaw-runtime-probe",
		Name:        "PicoClaw Runtime Probe",
		Description: "Read-only inspection of discovered services and processes.",
		ReadOnly:    true,
		Targets:     []string{"services", "processes"},
	}
}

func (runtimeProbeTool) Execute(ctx context.Context, report model.Report, input map[string]any) (any, error) {
	return nil, nil
}
