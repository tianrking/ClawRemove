package cline

import (
	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/skills"
	"github.com/tianrking/ClawRemove/internal/tools"
	"github.com/tianrking/ClawRemove/internal/verify"
)

type Provider struct{}

func (Provider) ID() string {
	return "cline"
}

func (Provider) DisplayName() string {
	return "Cline"
}

func (Provider) Facts() model.ProductFacts {
	return model.ProductFacts{
		ID:          "cline",
		DisplayName: "Cline",
		StateDirNames: []string{
			".cline",
			".cline_rules",
		},
		WorkspaceDirNames: []string{
			"tasks",
			"history",
		},
		ConfigNames: []string{
			"cline_mcp_settings.json",
			"settings.json",
			"api_config.json",
		},
		Markers: []string{
			"cline",
			"Cline",
			"saoudrizwan.claude-dev",
			"claude-dev",
		},
		ShellProfileGlobs: []string{
			".zshrc",
			".bashrc",
			".bash_profile",
			".profile",
			".config/fish/config.fish",
		},
		TempPrefixes: []string{
			"cline",
			"cline-",
			".cline-",
		},
		AppPaths: []string{
			".vscode/extensions/saoudrizwan.claude-dev",
			".cursor/extensions/saoudrizwan.claude-dev",
			".windsurf/extensions/saoudrizwan.claude-dev",
			"Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev",
			"AppData/Roaming/Code/User/globalStorage/saoudrizwan.claude-dev",
			".config/Code/User/globalStorage/saoudrizwan.claude-dev",
		},
		CLIPaths: []string{},
		PackageRefs: []model.PackageRef{
			{Manager: "vscode", Name: "saoudrizwan.claude-dev", Kind: "extension"},
		},
		ListenerPorts: []int{},
		RegistryPaths: []string{},
		EnvVarNames: []string{
			"CLINE_API_KEY",
			"ANTHROPIC_API_KEY",
			"OPENAI_API_KEY",
			"OPENROUTER_API_KEY",
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
