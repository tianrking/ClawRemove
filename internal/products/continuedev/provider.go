package continuedev

import (
	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/skills"
	"github.com/tianrking/ClawRemove/internal/tools"
	"github.com/tianrking/ClawRemove/internal/verify"
)

type Provider struct{}

func (Provider) ID() string {
	return "continue"
}

func (Provider) DisplayName() string {
	return "Continue"
}

func (Provider) Facts() model.ProductFacts {
	return model.ProductFacts{
		ID:          "continue",
		DisplayName: "Continue",
		StateDirNames: []string{
			".continue",
			".continue-dev",
		},
		WorkspaceDirNames: []string{
			"sessions",
			"logs",
			"embeddings",
		},
		ConfigNames: []string{
			"config.json",
			"config.ts",
			"continue.json",
		},
		Markers: []string{
			"continue",
			"Continue",
			"continue-dev",
			"continuedev",
		},
		ShellProfileGlobs: []string{
			".zshrc",
			".bashrc",
			".bash_profile",
			".profile",
			".config/fish/config.fish",
		},
		TempPrefixes: []string{
			"continue",
			"continue-",
			".continue-",
		},
		AppPaths: []string{
			".vscode/extensions/continue.continue",
			".cursor/extensions/continue.continue",
			".windsurf/extensions/continue.continue",
			"Library/Application Support/Code/User/globalStorage/continue.continue",
			"AppData/Roaming/Code/User/globalStorage/continue.continue",
			".config/Code/User/globalStorage/continue.continue",
			".local/share/JetBrains/Continue",
			"Library/Application Support/JetBrains/Continue",
		},
		CLIPaths: []string{
			".continue/bin/continue",
		},
		PackageRefs: []model.PackageRef{
			{Manager: "vscode", Name: "continue.continue", Kind: "extension"},
			{Manager: "npm", Name: "@continuedev/continue", Kind: "package"},
		},
		ListenerPorts: []int{
			6543, // Continue default port for local LLM
		},
		RegistryPaths: []string{},
		EnvVarNames: []string{
			"CONTINUE_API_KEY",
			"CONTINUE_CONFIG_DIR",
			"OPENAI_API_KEY",
			"ANTHROPIC_API_KEY",
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
