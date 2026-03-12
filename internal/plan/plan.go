package plan

import (
	"path/filepath"
	"runtime"

	"github.com/tianrking/ClawRemove/internal/model"
)

func Build(discovery model.Discovery, evidence model.EvidenceSet, facts model.ProductFacts, options model.Options) model.Plan {
	var actions []model.Action

	if !options.AuditOnly {
		actions = append(actions, serviceActions(discovery, evidence, facts, options)...)
		actions = append(actions, packageActions(discovery, evidence, facts, options)...)
		actions = append(actions, shellActions(discovery, options)...)
		actions = append(actions, processActions(discovery, evidence, options)...)
		actions = append(actions, containerActions(discovery, evidence, options)...)
		actions = append(actions, pathActions(discovery, evidence, facts, options)...)
	}

	actions = append(actions, auditActions(evidence)...)
	return model.Plan{Actions: actions}
}

func serviceActions(discovery model.Discovery, evidence model.EvidenceSet, facts model.ProductFacts, options model.Options) []model.Action {
	var actions []model.Action
	for _, service := range discovery.Services {
		ev := matchEvidence(evidence, "service", service.Path, service.Platform+":"+service.Name)
		switch service.Platform {
		case "darwin":
			actions = append(actions, model.Action{
				Kind:     model.ActionRunCommand,
				Target:   service.Name,
				Reason:   "Unload " + facts.DisplayName + " launchd service",
				Evidence: ev.Strength,
				Command:  []string{"launchctl", "bootout", "gui/$UID/" + service.Name},
				Platform: service.Platform,
			})
		case "linux":
			args := []string{"systemctl"}
			if service.Scope == "user" {
				args = append(args, "--user")
			}
			args = append(args, "disable", "--now", service.Name+".service")
			actions = append(actions, model.Action{
				Kind:     model.ActionRunCommand,
				Target:   service.Name,
				Reason:   "Disable " + facts.DisplayName + " systemd service",
				Evidence: ev.Strength,
				Command:  args,
				Platform: service.Platform,
			})
		case "windows":
			actions = append(actions, model.Action{
				Kind:     model.ActionRunCommand,
				Target:   service.Name,
				Reason:   "Delete " + facts.DisplayName + " scheduled task",
				Evidence: ev.Strength,
				Command:  []string{"schtasks", "/Delete", "/F", "/TN", service.Name},
				Platform: service.Platform,
			})
		}
		if service.Path != "" {
			actions = append(actions, model.Action{
				Kind:     model.ActionRemovePath,
				Target:   service.Path,
				Reason:   "Remove " + facts.DisplayName + " service definition file",
				Evidence: ev.Strength,
				Platform: service.Platform,
			})
		}
	}
	return actions
}

func packageActions(discovery model.Discovery, evidence model.EvidenceSet, facts model.ProductFacts, options model.Options) []model.Action {
	if options.KeepCLI {
		return nil
	}
	var actions []model.Action
	for _, pkg := range discovery.Packages {
		ev := matchEvidence(evidence, "package", pkg.Manager+":"+pkg.Name)
		cmd := []string{pkg.Manager}
		switch pkg.Manager {
		case "npm":
			cmd = append(cmd, "uninstall", "-g", pkg.Name)
		case "pnpm":
			cmd = append(cmd, "remove", "-g", pkg.Name)
		case "bun":
			cmd = append(cmd, "remove", "-g", pkg.Name)
		case "brew":
			cmd = append(cmd, "uninstall")
			if pkg.Kind == "cask" {
				cmd = append(cmd, "--cask")
			}
			cmd = append(cmd, pkg.Name)
		default:
			continue
		}
		actions = append(actions, model.Action{
			Kind:     model.ActionRunCommand,
			Target:   pkg.Name,
			Reason:   "Remove installed " + facts.DisplayName + " package",
			Evidence: ev.Strength,
			Command:  cmd,
			Platform: runtime.GOOS,
		})
	}
	return actions
}

func shellActions(discovery model.Discovery, options model.Options) []model.Action {
	if options.KeepShell {
		return nil
	}
	var actions []model.Action
	for _, file := range discovery.ShellProfiles {
		actions = append(actions, model.Action{
			Kind:   model.ActionEditFile,
			Target: file,
			Reason: "Remove OpenClaw completion block",
		})
	}
	return actions
}

func processActions(discovery model.Discovery, evidence model.EvidenceSet, options model.Options) []model.Action {
	if !options.KillProcesses {
		return nil
	}
	var actions []model.Action
	for _, proc := range discovery.Processes {
		if proc.PID == 0 {
			continue
		}
		ev := matchEvidence(evidence, "process", proc.Command)
		command := []string{"kill", "-TERM", itoa(proc.PID)}
		if runtime.GOOS == "windows" {
			command = []string{"taskkill", "/PID", itoa(proc.PID), "/F"}
		}
		actions = append(actions, model.Action{
			Kind:     model.ActionRunCommand,
			Target:   proc.Command,
			Reason:   "Terminate matching process",
			Evidence: ev.Strength,
			Command:  command,
			HighRisk: true,
			Risk:     "Kills live process",
		})
	}
	return actions
}

func containerActions(discovery model.Discovery, evidence model.EvidenceSet, options model.Options) []model.Action {
	var actions []model.Action
	for _, container := range discovery.Containers {
		ev := matchEvidence(evidence, "container", container.Runtime+":"+container.ID)
		if !options.RemoveDocker {
			actions = append(actions, model.Action{
				Kind:     model.ActionReportOnly,
				Target:   container.Runtime + ":" + container.ID,
				Reason:   "Matching container detected; enable --remove-docker to delete",
				Evidence: ev.Strength,
				HighRisk: true,
				Risk:     "Deletes container state",
			})
			continue
		}
		actions = append(actions, model.Action{
			Kind:     model.ActionRunCommand,
			Target:   container.Runtime + ":" + container.ID,
			Reason:   "Remove matching container",
			Evidence: ev.Strength,
			Command:  []string{container.Runtime, "rm", "-f", container.ID},
			HighRisk: true,
			Risk:     "Deletes container state",
		})
	}
	for _, image := range discovery.Images {
		name := image.Name
		if image.ID != "" {
			name = image.ID
		}
		ev := matchEvidence(evidence, "image", image.Runtime+":"+name)
		if !options.RemoveDocker {
			actions = append(actions, model.Action{
				Kind:     model.ActionReportOnly,
				Target:   image.Runtime + ":" + image.Name,
				Reason:   "Matching image detected; enable --remove-docker to delete",
				Evidence: ev.Strength,
				HighRisk: true,
				Risk:     "Deletes local image",
			})
			continue
		}
		target := image.ID
		if target == "" {
			target = image.Name
		}
		actions = append(actions, model.Action{
			Kind:     model.ActionRunCommand,
			Target:   image.Runtime + ":" + target,
			Reason:   "Remove matching image",
			Evidence: ev.Strength,
			Command:  []string{image.Runtime, "rmi", "-f", target},
			HighRisk: true,
			Risk:     "Deletes local image",
		})
	}
	return actions
}

func pathActions(discovery model.Discovery, evidence model.EvidenceSet, facts model.ProductFacts, options model.Options) []model.Action {
	var out []model.Action
	appendRemove := func(kind string, paths []string, reason string) {
		for _, item := range paths {
			ev := matchEvidence(evidence, kind, item)
			out = append(out, model.Action{
				Kind:     model.ActionRemovePath,
				Target:   item,
				Reason:   reason,
				Evidence: ev.Strength,
			})
		}
	}

	appendRemove("state_dir", discovery.StateDirs, "Remove "+facts.DisplayName+" state directory")
	if !options.KeepWorkspace {
		appendRemove("workspace_dir", discovery.WorkspaceDirs, "Remove "+facts.DisplayName+" workspace")
	}
	appendRemove("temp_path", discovery.TempPaths, "Remove "+facts.DisplayName+" temp path")
	if !options.KeepCLI {
		appendRemove("cli_path", discovery.CLIPaths, "Remove "+facts.DisplayName+" CLI wrapper")
	}
	if !options.KeepApp {
		appendRemove("app_path", discovery.AppPaths, "Remove "+facts.DisplayName+" app artifact")
	}

	return dedupeActions(out)
}

func auditActions(evidence model.EvidenceSet) []model.Action {
	var actions []model.Action
	for _, item := range evidence.Items {
		if item.Kind != "listener" && item.Kind != "crontab" && item.Kind != "image" && item.Kind != "shell_profile" {
			continue
		}
		actions = append(actions, model.Action{
			Kind:     model.ActionReportOnly,
			Target:   item.Target,
			Reason:   item.Reason,
			Evidence: item.Strength,
			HighRisk: item.Risk == "high",
			Risk:     item.Risk,
		})
	}
	return actions
}

func matchEvidence(evidence model.EvidenceSet, kind string, targets ...string) model.Evidence {
	for _, item := range evidence.Items {
		if item.Kind != kind {
			continue
		}
		for _, target := range targets {
			if target != "" && item.Target == target {
				return item
			}
		}
	}
	return model.Evidence{}
}

func dedupeActions(actions []model.Action) []model.Action {
	seen := map[string]struct{}{}
	var out []model.Action
	for _, action := range actions {
		key := string(action.Kind) + "|" + action.Target + "|" + filepath.Clean(action.Reason)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, action)
	}
	return out
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	var buf [32]byte
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + (value % 10))
		value /= 10
	}
	return sign + string(buf[i:])
}
