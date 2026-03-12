package aider

import (
	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/skills"
	"github.com/tianrking/ClawRemove/internal/tools"
	"github.com/tianrking/ClawRemove/internal/verify"
)

type Provider struct{}

func (Provider) ID() string {
	return "aider"
}

func (Provider) DisplayName() string {
	return "Aider"
}

func (Provider) Facts() model.ProductFacts {
	return model.ProductFacts{
		ID:          "aider",
		DisplayName: "Aider",
		StateDirNames: []string{
			".aider",
			".aider.chat.history",
		},
		WorkspaceDirNames: []string{},
		ConfigNames: []string{
			"aider.conf.yml",
			".aider.conf.yml",
		},
		Markers: []string{
			"aider",
			"Aider",
			"aider-ai",
			"aider-chat",
		},
		ShellProfileGlobs: []string{
			".zshrc",
			".bashrc",
			".bash_profile",
			".profile",
		},
		TempPrefixes: []string{
			"aider",
			"aider-",
			".aider-",
		},
		AppPaths: []string{},
		CLIPaths: []string{
			".local/bin/aider",
		},
		PackageRefs: []model.PackageRef{
			{Manager: "pip", Name: "aider-chat"},
			{Manager: "pipx", Name: "aider-chat"},
			{Manager: "brew", Name: "aider", Kind: "formula"},
		},
		ListenerPorts: []int{},
		RegistryPaths: []string{},
		EnvVarNames: []string{
			"AIDER_API_KEY",
			"AIDER_OPENAI_API_KEY",
			"AIDER_ANTHROPIC_API_KEY",
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
