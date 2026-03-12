package zeroclaw

import (
	"context"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/skills"
	"github.com/tianrking/ClawRemove/internal/tools"
	"github.com/tianrking/ClawRemove/internal/verify"
)

type Provider struct{}

func (Provider) ID() string {
	return "zeroclaw"
}

func (Provider) DisplayName() string {
	return "ZeroClaw"
}

func (Provider) Facts() model.ProductFacts {
	return model.ProductFacts{
		ID:          "zeroclaw",
		DisplayName: "ZeroClaw",
		StateDirNames: []string{
			".zeroclaw",
			".zero",
		},
		WorkspaceDirNames: []string{
			"workspace",
			"sessions",
			"memory",
		},
		ConfigNames: []string{
			"zeroclaw.json",
			"zero.json",
		},
		Markers: []string{
			"zeroclaw",
			"zero.claw",
			"ai.zeroclaw",
			"com.zeroclaw",
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
			"zeroclaw",
			"zeroclaw-",
		},
		AppPaths: []string{
			"/Applications/ZeroClaw.app",
			"Applications/ZeroClaw.app",
			"Library/Application Support/ZeroClaw",
			"Library/Caches/ZeroClaw",
			".local/share/applications/zeroclaw.desktop",
			".config/ZeroClaw",
			"AppData/Roaming/ZeroClaw",
			"AppData/Local/Programs/ZeroClaw",
		},
		CLIPaths: []string{
			".local/bin/zeroclaw",
		},
		PackageRefs: []model.PackageRef{
			{Manager: "npm", Name: "zeroclaw"},
			{Manager: "pnpm", Name: "zeroclaw"},
			{Manager: "bun", Name: "zeroclaw"},
		},
		ListenerPorts: []int{18792, 19004},
		RegistryPaths: []string{
			"HKCU\\Software\\ZeroClaw",
			"HKLM\\SOFTWARE\\ZeroClaw",
		},
		EnvVarNames: []string{
			"ZEROCLAW_HOME",
			"ZEROCLAW_CONFIG",
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
		ID:          "zeroclaw-residue-analysis",
		Name:        "ZeroClaw Residue Analysis",
		Description: "Analyze ZeroClaw residue using provider-specific state paths and markers.",
		Inputs:      []string{"discovery", "verification"},
	}
}

func (residueAnalysisSkill) Analyze(ctx context.Context, report model.Report) (any, error) {
	return nil, nil
}

type safeRemovalReviewSkill struct{}

func (safeRemovalReviewSkill) Info() model.ProviderSkill {
	return model.ProviderSkill{
		ID:          "zeroclaw-safe-removal-review",
		Name:        "ZeroClaw Safe Removal Review",
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
		ID:          "zeroclaw-state-probe",
		Name:        "ZeroClaw State Probe",
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
		ID:          "zeroclaw-runtime-probe",
		Name:        "ZeroClaw Runtime Probe",
		Description: "Read-only inspection of packages and processes.",
		ReadOnly:    true,
		Targets:     []string{"packages", "processes"},
	}
}

func (runtimeProbeTool) Execute(ctx context.Context, report model.Report, input map[string]any) (any, error) {
	return nil, nil
}
