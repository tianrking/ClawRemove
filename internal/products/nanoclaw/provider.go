package nanoclaw

import (
	"context"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/skills"
	"github.com/tianrking/ClawRemove/internal/tools"
	"github.com/tianrking/ClawRemove/internal/verify"
)

type Provider struct{}

func (Provider) ID() string {
	return "nanoclaw"
}

func (Provider) DisplayName() string {
	return "NanoClaw"
}

func (Provider) Facts() model.ProductFacts {
	return model.ProductFacts{
		ID:          "nanoclaw",
		DisplayName: "NanoClaw",
		StateDirNames: []string{
			".nanoclaw",
			".nano",
		},
		WorkspaceDirNames: []string{
			"workspace",
			"projects",
			"history",
		},
		ConfigNames: []string{
			"nanoclaw.json",
			"nano.json",
		},
		Markers: []string{
			"nanoclaw",
			"nano.claw",
			"ai.nanoclaw",
			"com.nanoclaw",
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
			"nanoclaw",
			"nanoclaw-",
		},
		AppPaths: []string{
			"/Applications/NanoClaw.app",
			"Applications/NanoClaw.app",
			"Library/Application Support/NanoClaw",
			"Library/Caches/NanoClaw",
			".local/share/applications/nanoclaw.desktop",
			".config/NanoClaw",
			"AppData/Roaming/NanoClaw",
			"AppData/Local/Programs/NanoClaw",
		},
		CLIPaths: []string{
			".local/bin/nanoclaw",
		},
		PackageRefs: []model.PackageRef{
			{Manager: "npm", Name: "nanoclaw"},
			{Manager: "pip", Name: "nanoclaw"},
		},
		ListenerPorts: []int{18793, 19005},
		RegistryPaths: []string{
			"HKCU\\Software\\NanoClaw",
			"HKLM\\SOFTWARE\\NanoClaw",
		},
		EnvVarNames: []string{
			"NANOCLAW_HOME",
			"NANOCLAW_CONFIG",
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
		ID:          "nanoclaw-residue-analysis",
		Name:        "NanoClaw Residue Analysis",
		Description: "Analyze NanoClaw residue using provider-specific state paths and markers.",
		Inputs:      []string{"discovery", "verification"},
	}
}

func (residueAnalysisSkill) Analyze(ctx context.Context, report model.Report) (any, error) {
	return nil, nil
}

type safeRemovalReviewSkill struct{}

func (safeRemovalReviewSkill) Info() model.ProviderSkill {
	return model.ProviderSkill{
		ID:          "nanoclaw-safe-removal-review",
		Name:        "NanoClaw Safe Removal Review",
		Description: "Review confirmed residue before apply.",
		Inputs:      []string{"verification", "plan", "advice"},
	}
}

func (safeRemovalReviewSkill) Analyze(ctx context.Context, report model.Report) (any, error) {
	return nil, nil
}

type stateProbeTool struct{}

func (stateProbeTool) Info() model.ProviderTool {
	return model.ProviderTool{
		ID:          "nanoclaw-state-probe",
		Name:        "NanoClaw State Probe",
		Description: "Read-only inspection of discovered state paths.",
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
		ID:          "nanoclaw-runtime-probe",
		Name:        "NanoClaw Runtime Probe",
		Description: "Read-only inspection of packages and processes.",
		ReadOnly:    true,
		Targets:     []string{"packages", "processes"},
	}
}

func (runtimeProbeTool) Execute(ctx context.Context, report model.Report, input map[string]any) (any, error) {
	return nil, nil
}
