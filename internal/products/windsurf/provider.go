package windsurf

import (
	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/skills"
	"github.com/tianrking/ClawRemove/internal/tools"
	"github.com/tianrking/ClawRemove/internal/verify"
)

type Provider struct{}

func (Provider) ID() string {
	return "windsurf"
}

func (Provider) DisplayName() string {
	return "Windsurf"
}

func (Provider) Facts() model.ProductFacts {
	return model.ProductFacts{
		ID:          "windsurf",
		DisplayName: "Windsurf",
		StateDirNames: []string{
			".windsurf",
			".windsurf-server",
		},
		WorkspaceDirNames: []string{
			"workspace",
			"projects",
		},
		ConfigNames: []string{
			"windsurf.json",
			"settings.json",
		},
		Markers: []string{
			"windsurf",
			"Windsurf",
			"codeium",
			"Codeium",
		},
		ShellProfileGlobs: []string{
			".zshrc",
			".bashrc",
			".bash_profile",
			".profile",
			".config/fish/config.fish",
		},
		TempPrefixes: []string{
			"windsurf",
			"windsurf-",
			".windsurf-",
		},
		AppPaths: []string{
			"/Applications/Windsurf.app",
			"Applications/Windsurf.app",
			"Library/Application Support/Windsurf",
			"Library/Caches/Windsurf",
			"Library/Preferences/com.exafunction.windsurf.plist",
			"Library/Saved Application State/com.exafunction.windsurf.savedState",
			".local/share/applications/windsurf.desktop",
			".config/Windsurf",
			"AppData/Roaming/Windsurf",
			"AppData/Local/Programs/Windsurf",
		},
		CLIPaths: []string{
			".local/bin/windsurf",
		},
		PackageRefs: []model.PackageRef{},
		ListenerPorts: []int{},
		RegistryPaths: []string{
			"HKCU\\Software\\Windsurf",
			"HKLM\\SOFTWARE\\Windsurf",
		},
		EnvVarNames: []string{
			"WINDSURF_API_KEY",
			"CODEIUM_API_KEY",
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
	return []skills.Skill{}
}

func (Provider) VerificationRules() []verify.Rule {
	return []verify.Rule{}
}

func (Provider) Tools() []tools.Tool {
	return []tools.Tool{}
}
