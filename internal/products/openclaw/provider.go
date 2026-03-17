package openclaw

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/skills"
	"github.com/tianrking/ClawRemove/internal/tools"
	"github.com/tianrking/ClawRemove/internal/verify"
)

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
			".openclaw-legacy",
			".claw-dev",
		},
		WorkspaceDirNames: []string{
			"workspace",
			"workspaces",
			"conversations",
			"projects",
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
			"openclaw-legacy",
			"openclaw-beta",
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
			"openclaw-updater-",
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
			".local/share/applications/openclaw.desktop",
			".config/OpenClaw",
			"AppData/Roaming/OpenClaw",
			"AppData/Local/Programs/OpenClaw",
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
		// OpenClaw's known gateway and IPC ports
		ListenerPorts: []int{18789, 19001, 19002},
		// Windows registry paths
		RegistryPaths: []string{
			"HKCU\\Software\\OpenClaw",
			"HKCU\\Software\\ai.openclaw",
			"HKLM\\SOFTWARE\\OpenClaw",
			"HKLM\\SOFTWARE\\WOW6432Node\\OpenClaw",
			"HKLM\\SOFTWARE\\ai.openclaw",
		},
		// Environment variables
		EnvVarNames: []string{
			"OPENCLAW_PATH",
			"OPENCLAW_HOME",
			"OPENCLAW_CONFIG",
			"OPENCLAW_API_KEY",
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
	return []verify.Rule{
		shellProfileVerificationRule{},
	}
}

func (Provider) Tools() []tools.Tool {
	return []tools.Tool{
		stateProbeTool{},
		runtimeProbeTool{},
		shellProbeTool{},
	}
}

type shellProfileVerificationRule struct{}

func (shellProfileVerificationRule) Evaluate(residual *model.Residual) {
	if residual.Kind == "shell_profile" {
		// OpenClaw modifies shell profiles distinctly with its own exact CLI paths and variables.
		// If we matched the profile due to an exact marker (like .openclaw), we can trust it is strong evidence.
		residual.Evidence = "strong"
		residual.Rule = "verified-shell-profile"
		residual.Confidence = 0.85
	}
}

type residueAnalysisSkill struct{}

func (residueAnalysisSkill) Info() model.ProviderSkill {
	return model.ProviderSkill{
		ID:          "openclaw-residue-analysis",
		Name:        "OpenClaw Residue Analysis",
		Description: "Analyze OpenClaw residue using provider-specific state paths, package names, service markers, and legacy aliases.",
		Inputs:      []string{"discovery", "verification"},
	}
}

func (residueAnalysisSkill) Analyze(ctx context.Context, report model.Report) (any, error) {
	// For now, this just acts as a structural contract for the LLM.
	return nil, nil
}

type safeRemovalReviewSkill struct{}

func (safeRemovalReviewSkill) Info() model.ProviderSkill {
	return model.ProviderSkill{
		ID:          "openclaw-safe-removal-review",
		Name:        "OpenClaw Safe Removal Review",
		Description: "Review confirmed residue, high-risk actions, and investigate-only findings before apply.",
		Inputs:      []string{"verification", "plan", "advice"},
	}
}

func (safeRemovalReviewSkill) Analyze(ctx context.Context, report model.Report) (any, error) {
	return nil, nil
}

type stateProbeTool struct{}

func (stateProbeTool) Info() model.ProviderTool {
	return model.ProviderTool{
		ID:          "openclaw-state-probe",
		Name:        "OpenClaw State Probe",
		Description: "Read-only inspection of discovered state, workspace, temp, app, and CLI paths.",
		ReadOnly:    true,
		Targets:     []string{"state_dirs", "workspace_dirs", "path_probe"},
	}
}

func (stateProbeTool) Execute(ctx context.Context, report model.Report, input map[string]any) (any, error) {
	targetRaw, ok := input["target"]
	if !ok {
		return map[string]any{
			"stateDirs":     report.Discovery.StateDirs,
			"workspaceDirs": report.Discovery.WorkspaceDirs,
			"tempPaths":     report.Discovery.TempPaths,
			"appPaths":      report.Discovery.AppPaths,
			"cliPaths":      report.Discovery.CLIPaths,
		}, nil
	}
	target, ok := targetRaw.(string)
	if !ok {
		return nil, fmt.Errorf("target must be string")
	}

	info, err := os.Stat(target)
	if err != nil {
		return map[string]any{"target": target, "exists": false}, nil
	}

	result := map[string]any{
		"target": target,
		"exists": true,
		"isDir":  info.IsDir(),
		"size":   info.Size(),
	}

	if info.IsDir() {
		entries, _ := os.ReadDir(target)
		var names []string
		for i, entry := range entries {
			if i >= 20 {
				break
			}
			names = append(names, entry.Name())
		}
		result["entries"] = names
	}

	return result, nil
}

type runtimeProbeTool struct{}

func (runtimeProbeTool) Info() model.ProviderTool {
	return model.ProviderTool{
		ID:          "openclaw-runtime-probe",
		Name:        "OpenClaw Runtime Probe",
		Description: "Read-only inspection of discovered packages, services, listeners, and processes.",
		ReadOnly:    true,
		Targets:     []string{"services", "service_probe", "packages", "package_probe", "processes", "process_probe", "verification"},
	}
}

func (runtimeProbeTool) Execute(ctx context.Context, report model.Report, input map[string]any) (any, error) {
	targetRaw, ok := input["target"]
	if !ok {
		return map[string]any{
			"services":     report.Discovery.Services,
			"packages":     report.Discovery.Packages,
			"processes":    report.Discovery.Processes,
			"listeners":    report.Discovery.Listeners,
			"registryKeys": report.Discovery.RegistryKeys,
		}, nil
	}

	target, _ := targetRaw.(string)
	for _, svc := range report.Discovery.Services {
		if svc.Name == target || svc.Path == target {
			result := map[string]any{"type": "service", "name": svc.Name, "platform": svc.Platform}
			if svc.Path != "" {
				if content, err := os.ReadFile(svc.Path); err == nil {
					lines := strings.Split(string(content), "\n")
					if len(lines) > 20 {
						lines = lines[:20]
					}
					result["preview"] = lines
				}
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("runtime target %q not found", target)
}

type shellProbeTool struct{}

func (shellProbeTool) Info() model.ProviderTool {
	return model.ProviderTool{
		ID:          "openclaw-shell-probe",
		Name:        "OpenClaw Shell Probe",
		Description: "Read-only inspection of shell profile traces and completion residue tied to OpenClaw markers.",
		ReadOnly:    true,
		Targets:     []string{"shell_profile_probe"},
	}
}

func (shellProbeTool) Execute(ctx context.Context, report model.Report, input map[string]any) (any, error) {
	targetRaw, ok := input["target"]
	if !ok {
		return map[string]any{
			"shellProfiles": report.Discovery.ShellProfiles,
		}, nil
	}

	target, _ := targetRaw.(string)
	for _, profile := range report.Discovery.ShellProfiles {
		if profile == target {
			content, err := os.ReadFile(target)
			if err != nil {
				return map[string]any{"exists": false, "error": err.Error()}, nil
			}
			
			var markers []string
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				lower := strings.ToLower(line)
				if strings.Contains(lower, "openclaw") || strings.Contains(lower, "clawdbot") {
					markers = append(markers, strings.TrimSpace(line))
				}
			}
			
			return map[string]any{
				"target":  target,
				"exists":  true,
				"matches": markers,
				"count":   len(markers),
			}, nil
		}
	}
	
	return nil, fmt.Errorf("shell profile %q not found in discovery", target)
}
