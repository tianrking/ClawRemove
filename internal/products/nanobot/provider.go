package nanobot

import (
	"context"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/skills"
	"github.com/tianrking/ClawRemove/internal/tools"
	"github.com/tianrking/ClawRemove/internal/verify"
)

type Provider struct{}

func (Provider) ID() string {
	return "nanobot"
}

func (Provider) DisplayName() string {
	return "NanoBot"
}

func (Provider) Facts() model.ProductFacts {
	return model.ProductFacts{
		ID:          "nanobot",
		DisplayName: "NanoBot",
		StateDirNames: []string{
			".nanobot",
		},
		WorkspaceDirNames: []string{
			"workspace",
			"history",
			"bridge",
			"sessions",
			"cron",
			"media",
			"logs",
			"whatsapp-auth",
		},
		ConfigNames: []string{
			"config.json",
		},
		Markers: []string{
			"nanobot",
			"nanobot-ai",
			"bot.nano",
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
			"nanobot",
			"nanobot-",
		},
		AppPaths: []string{
			".local/share/nanobot",
			".config/nanobot",
			"Library/Application Support/NanoBot",
			"Library/Caches/NanoBot",
			"AppData/Roaming/NanoBot",
			"AppData/Local/NanoBot",
		},
		CLIPaths: []string{
			".local/bin/nanobot",
		},
		PackageRefs: []model.PackageRef{
			{Manager: "pip", Name: "nanobot-ai"},
			{Manager: "pipx", Name: "nanobot-ai"},
		},
		// NanoBot's known gateway and IPC ports
		ListenerPorts: []int{18790, 3001},
		// Windows registry paths
		RegistryPaths: []string{
			"HKCU\\Software\\NanoBot",
			"HKLM\\SOFTWARE\\NanoBot",
		},
		// Environment variables
		EnvVarNames: []string{
			"NANOBOT_TMUX_SOCKET_DIR",
			"NANOBOT_HOME",
			"NANOBOT_CONFIG",
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
		ID:          "nanobot-residue-analysis",
		Name:        "NanoBot Residue Analysis",
		Description: "Analyze NanoBot residue using provider-specific state paths, package names, and service markers.",
		Inputs:      []string{"discovery", "verification"},
	}
}

func (residueAnalysisSkill) Analyze(ctx context.Context, report model.Report) (any, error) {
	return nil, nil
}

type safeRemovalReviewSkill struct{}

func (safeRemovalReviewSkill) Info() model.ProviderSkill {
	return model.ProviderSkill{
		ID:          "nanobot-safe-removal-review",
		Name:        "NanoBot Safe Removal Review",
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
		ID:          "nanobot-state-probe",
		Name:        "NanoBot State Probe",
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
		ID:          "nanobot-runtime-probe",
		Name:        "NanoBot Runtime Probe",
		Description: "Read-only inspection of discovered packages, services, and processes.",
		ReadOnly:    true,
		Targets:     []string{"services", "packages", "processes"},
	}
}

func (runtimeProbeTool) Execute(ctx context.Context, report model.Report, input map[string]any) (any, error) {
	return nil, nil
}
