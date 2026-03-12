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
