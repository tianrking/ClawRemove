package cursor

import (
	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/skills"
	"github.com/tianrking/ClawRemove/internal/tools"
	"github.com/tianrking/ClawRemove/internal/verify"
)

type Provider struct{}

func (Provider) ID() string {
	return "cursor"
}

func (Provider) DisplayName() string {
	return "Cursor"
}

func (Provider) Facts() model.ProductFacts {
	return model.ProductFacts{
		ID:          "cursor",
		DisplayName: "Cursor",
		StateDirNames: []string{
			".cursor",
			".cursor-tutor",
		},
		WorkspaceDirNames: []string{
			"workspace",
			"projects",
		},
		ConfigNames: []string{
			"cursor.json",
			"settings.json",
			"argv.json",
		},
		Markers: []string{
			"cursor",
			"Cursor",
			"cursor-ai",
			"getcursor",
		},
		ShellProfileGlobs: []string{
			".zshrc",
			".bashrc",
			".bash_profile",
			".profile",
			".config/fish/config.fish",
		},
		TempPrefixes: []string{
			"cursor",
			"cursor-",
			".cursor-",
		},
		AppPaths: []string{
			"/Applications/Cursor.app",
			"Applications/Cursor.app",
			"Library/Application Support/Cursor",
			"Library/Caches/Cursor",
			"Library/Preferences/com.todesktop.230313mzl4w4u92.plist",
			"Library/Saved Application State/com.todesktop.230313mzl4w4u92.savedState",
			".local/share/applications/cursor.desktop",
			".config/Cursor",
			"AppData/Roaming/Cursor",
			"AppData/Local/Programs/Cursor",
		},
		CLIPaths: []string{
			".local/bin/cursor",
		},
		PackageRefs: []model.PackageRef{
			{Manager: "brew", Name: "cursor", Kind: "cask"},
		},
		ListenerPorts: []int{},
		RegistryPaths: []string{
			"HKCU\\Software\\Cursor",
			"HKLM\\SOFTWARE\\Cursor",
		},
		EnvVarNames: []string{
			"CURSOR_API_KEY",
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
