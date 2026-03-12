package openfang

import (
	"context"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/skills"
	"github.com/tianrking/ClawRemove/internal/tools"
	"github.com/tianrking/ClawRemove/internal/verify"
)

type Provider struct{}

func (Provider) ID() string {
	return "openfang"
}

func (Provider) DisplayName() string {
	return "OpenFang"
}

func (Provider) Facts() model.ProductFacts {
	return model.ProductFacts{
		ID:          "openfang",
		DisplayName: "OpenFang",
		StateDirNames: []string{
			".openfang",
			".fang",
		},
		WorkspaceDirNames: []string{
			"workspace",
			"workspaces",
			"projects",
		},
		ConfigNames: []string{
			"openfang.json",
			"fang.json",
		},
		Markers: []string{
			"openfang",
			"open.fang",
			"ai.openfang",
			"com.openfang",
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
			"openfang",
			"openfang-",
		},
		AppPaths: []string{
			"/Applications/OpenFang.app",
			"Applications/OpenFang.app",
			"Library/Application Support/OpenFang",
			"Library/Caches/OpenFang",
			".local/share/applications/openfang.desktop",
			".config/OpenFang",
			"AppData/Roaming/OpenFang",
			"AppData/Local/Programs/OpenFang",
		},
		CLIPaths: []string{
			".local/bin/openfang",
		},
		PackageRefs: []model.PackageRef{
			{Manager: "npm", Name: "openfang"},
			{Manager: "pnpm", Name: "openfang"},
			{Manager: "bun", Name: "openfang"},
			{Manager: "brew", Name: "openfang", Kind: "formula"},
		},
		ListenerPorts: []int{18791, 19003},
		RegistryPaths: []string{
			"HKCU\\Software\\OpenFang",
			"HKLM\\SOFTWARE\\OpenFang",
		},
		EnvVarNames: []string{
			"OPENFANG_HOME",
			"OPENFANG_CONFIG",
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
		ID:          "openfang-residue-analysis",
		Name:        "OpenFang Residue Analysis",
		Description: "Analyze OpenFang residue using provider-specific state paths and markers.",
		Inputs:      []string{"discovery", "verification"},
	}
}

func (residueAnalysisSkill) Analyze(ctx context.Context, report model.Report) (any, error) {
	return nil, nil
}

type safeRemovalReviewSkill struct{}

func (safeRemovalReviewSkill) Info() model.ProviderSkill {
	return model.ProviderSkill{
		ID:          "openfang-safe-removal-review",
		Name:        "OpenFang Safe Removal Review",
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
		ID:          "openfang-state-probe",
		Name:        "OpenFang State Probe",
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
		ID:          "openfang-runtime-probe",
		Name:        "OpenFang Runtime Probe",
		Description: "Read-only inspection of packages and processes.",
		ReadOnly:    true,
		Targets:     []string{"packages", "processes"},
	}
}

func (runtimeProbeTool) Execute(ctx context.Context, report model.Report, input map[string]any) (any, error) {
	return nil, nil
}
