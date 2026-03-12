package plan

import (
	"path/filepath"
	"runtime"

	"github.com/tianrking/ClawRemove/internal/model"
)

func Build(discovery model.Discovery, facts model.ProductFacts, options model.Options) model.Plan {
	var actions []model.Action

	if !options.AuditOnly {
		actions = append(actions, serviceActions(discovery, facts, options)...)
		actions = append(actions, packageActions(discovery, facts, options)...)
		actions = append(actions, shellActions(discovery, options)...)
		actions = append(actions, processActions(discovery, options)...)
		actions = append(actions, containerActions(discovery, options)...)
		actions = append(actions, pathActions(discovery, facts, options)...)
	}

	actions = append(actions, auditActions(discovery)...)
	return model.Plan{Actions: actions}
}

func serviceActions(discovery model.Discovery, facts model.ProductFacts, options model.Options) []model.Action {
	var actions []model.Action
	for _, service := range discovery.Services {
		switch service.Platform {
		case "darwin":
			actions = append(actions, model.Action{
				Kind:     model.ActionRunCommand,
				Target:   service.Name,
				Reason:   "Unload " + facts.DisplayName + " launchd service",
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
				Command:  args,
				Platform: service.Platform,
			})
		case "windows":
			actions = append(actions, model.Action{
				Kind:     model.ActionRunCommand,
				Target:   service.Name,
				Reason:   "Delete " + facts.DisplayName + " scheduled task",
				Command:  []string{"schtasks", "/Delete", "/F", "/TN", service.Name},
				Platform: service.Platform,
			})
		}
		if service.Path != "" {
			actions = append(actions, model.Action{
				Kind:     model.ActionRemovePath,
				Target:   service.Path,
				Reason:   "Remove " + facts.DisplayName + " service definition file",
				Platform: service.Platform,
			})
		}
	}
	return actions
}

func packageActions(discovery model.Discovery, facts model.ProductFacts, options model.Options) []model.Action {
	if options.KeepCLI {
		return nil
	}
	var actions []model.Action
	for _, pkg := range discovery.Packages {
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

func processActions(discovery model.Discovery, options model.Options) []model.Action {
	if !options.KillProcesses {
		return nil
	}
	var actions []model.Action
	for _, proc := range discovery.Processes {
		if proc.PID == 0 {
			continue
		}
		command := []string{"kill", "-TERM", itoa(proc.PID)}
		if runtime.GOOS == "windows" {
			command = []string{"taskkill", "/PID", itoa(proc.PID), "/F"}
		}
		actions = append(actions, model.Action{
			Kind:     model.ActionRunCommand,
			Target:   proc.Command,
			Reason:   "Terminate matching process",
			Command:  command,
			HighRisk: true,
			Risk:     "Kills live process",
		})
	}
	return actions
}

func containerActions(discovery model.Discovery, options model.Options) []model.Action {
	var actions []model.Action
	for _, container := range discovery.Containers {
		if !options.RemoveDocker {
			actions = append(actions, model.Action{
				Kind:     model.ActionReportOnly,
				Target:   container.Runtime + ":" + container.ID,
				Reason:   "Matching container detected; enable --remove-docker to delete",
				HighRisk: true,
				Risk:     "Deletes container state",
			})
			continue
		}
		actions = append(actions, model.Action{
			Kind:     model.ActionRunCommand,
			Target:   container.Runtime + ":" + container.ID,
			Reason:   "Remove matching container",
			Command:  []string{container.Runtime, "rm", "-f", container.ID},
			HighRisk: true,
			Risk:     "Deletes container state",
		})
	}
	for _, image := range discovery.Images {
		if !options.RemoveDocker {
			actions = append(actions, model.Action{
				Kind:     model.ActionReportOnly,
				Target:   image.Runtime + ":" + image.Name,
				Reason:   "Matching image detected; enable --remove-docker to delete",
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
			Command:  []string{image.Runtime, "rmi", "-f", target},
			HighRisk: true,
			Risk:     "Deletes local image",
		})
	}
	return actions
}

func pathActions(discovery model.Discovery, facts model.ProductFacts, options model.Options) []model.Action {
	var out []model.Action
	appendRemove := func(paths []string, reason string) {
		for _, item := range paths {
			out = append(out, model.Action{
				Kind:   model.ActionRemovePath,
				Target: item,
				Reason: reason,
			})
		}
	}

	appendRemove(discovery.StateDirs, "Remove "+facts.DisplayName+" state directory")
	if !options.KeepWorkspace {
		appendRemove(discovery.WorkspaceDirs, "Remove "+facts.DisplayName+" workspace")
	}
	appendRemove(discovery.TempPaths, "Remove "+facts.DisplayName+" temp path")
	if !options.KeepCLI {
		appendRemove(discovery.CLIPaths, "Remove "+facts.DisplayName+" CLI wrapper")
	}
	if !options.KeepApp {
		appendRemove(discovery.AppPaths, "Remove "+facts.DisplayName+" app artifact")
	}

	return dedupeActions(out)
}

func auditActions(discovery model.Discovery) []model.Action {
	var actions []model.Action
	for _, line := range discovery.Listeners {
		actions = append(actions, model.Action{
			Kind:   model.ActionReportOnly,
			Target: line,
			Reason: "Listener matched OpenClaw markers or default ports",
		})
	}
	for _, line := range discovery.CrontabLines {
		actions = append(actions, model.Action{
			Kind:     model.ActionReportOnly,
			Target:   line,
			Reason:   "Crontab entry references OpenClaw markers",
			HighRisk: true,
			Risk:     "Manual review recommended before deletion",
		})
	}
	return actions
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
