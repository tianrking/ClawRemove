package openclaw

import "github.com/tianrking/ClawRemove/internal/model"

type Provider struct{}

func (Provider) ID() string {
	return "openclaw"
}

func (Provider) DisplayName() string {
	return "OpenClaw"
}

func (Provider) Facts() model.ProductFacts {
	return model.ProductFacts{
		ID:          "openclaw",
		DisplayName: "OpenClaw",
		StateDirNames: []string{
			".openclaw",
			".clawdbot",
			".moldbot",
			".moltbot",
		},
		ConfigNames: []string{
			"openclaw.json",
			"clawdbot.json",
			"moldbot.json",
			"moltbot.json",
		},
		Markers: []string{
			"openclaw",
			"ai.openclaw",
			"com.openclaw",
			"bot.molt",
			"clawdbot",
			"moltbot",
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
			"openclaw",
			"openclaw-",
			"openclaw-img-",
			"openclaw-restart-",
			"openclaw-zai-fallback-",
			".openclaw-install-stage-",
			".openclaw-install-backups",
		},
		AppPaths: []string{
			"/Applications/OpenClaw.app",
			"Applications/OpenClaw.app",
			"Library/Application Support/OpenClaw",
			"Library/Caches/OpenClaw",
			"Library/Preferences/ai.openclaw.mac.plist",
			"Library/Saved Application State/ai.openclaw.mac.savedState",
			"Library/WebKit/ai.openclaw.mac",
			"Library/HTTPStorages/ai.openclaw.mac",
			"Library/Cookies/ai.openclaw.mac.binarycookies",
			"Library/LaunchAgents/ai.openclaw.mac.plist",
		},
		CLIPaths: []string{
			".local/bin/openclaw",
			".local/bin/openclaw.cmd",
		},
		PackageRefs: []model.PackageRef{
			{Manager: "npm", Name: "openclaw"},
			{Manager: "pnpm", Name: "openclaw"},
			{Manager: "bun", Name: "openclaw"},
			{Manager: "brew", Name: "openclaw-cli", Kind: "formula"},
			{Manager: "brew", Name: "openclaw", Kind: "formula"},
			{Manager: "brew", Name: "openclaw", Kind: "cask"},
		},
	}
}

func (Provider) Capabilities() model.ProviderCapabilities {
	return model.ProviderCapabilities{
		Skills: []model.ProviderSkill{
			{
				ID:          "openclaw-residue-analysis",
				Name:        "OpenClaw Residue Analysis",
				Description: "Analyze OpenClaw residue using provider-specific state paths, package names, service markers, and legacy aliases.",
				Inputs:      []string{"discovery", "verification"},
			},
			{
				ID:          "openclaw-safe-removal-review",
				Name:        "OpenClaw Safe Removal Review",
				Description: "Review confirmed residue, high-risk actions, and investigate-only findings before apply.",
				Inputs:      []string{"verification", "plan", "advice"},
			},
		},
		Tools: []model.ProviderTool{
			{
				ID:          "openclaw-state-probe",
				Name:        "OpenClaw State Probe",
				Description: "Read-only inspection of discovered state, workspace, temp, app, and CLI paths.",
				ReadOnly:    true,
				Targets:     []string{"state_dirs", "workspace_dirs", "path_probe"},
			},
			{
				ID:          "openclaw-runtime-probe",
				Name:        "OpenClaw Runtime Probe",
				Description: "Read-only inspection of discovered packages, services, listeners, and processes.",
				ReadOnly:    true,
				Targets:     []string{"services", "service_probe", "packages", "package_probe", "processes", "process_probe", "verification"},
			},
			{
				ID:          "openclaw-shell-probe",
				Name:        "OpenClaw Shell Probe",
				Description: "Read-only inspection of shell profile traces and completion residue tied to OpenClaw markers.",
				ReadOnly:    true,
				Targets:     []string{"shell_profile_probe"},
			},
		},
	}
}
