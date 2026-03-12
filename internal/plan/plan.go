package plan

import (
	"path/filepath"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/platform"
)

const minConfidenceForDestructive = 0.7

func Build(discovery model.Discovery, evidence model.EvidenceSet, facts model.ProductFacts, options model.Options, host platform.Host) model.Plan {
	var actions []model.Action
	adapter := platform.NewAdapter(host)

	if !options.AuditOnly {
		actions = append(actions, serviceActions(discovery, evidence, facts, options, adapter)...)
		actions = append(actions, packageActions(discovery, evidence, facts, options)...)
		actions = append(actions, shellActions(discovery, options)...)
		actions = append(actions, processActions(discovery, evidence, options, adapter)...)
		actions = append(actions, containerActions(discovery, evidence, options)...)
		actions = append(actions, pathActions(discovery, evidence, facts, options)...)
	}

	actions = append(actions, auditActions(evidence)...)
	return model.Plan{Actions: actions}
}

func serviceActions(discovery model.Discovery, evidence model.EvidenceSet, facts model.ProductFacts, options model.Options, adapter platform.Adapter) []model.Action {
	_ = options
	var actions []model.Action
	for _, service := range discovery.Services {
		ev := matchEvidence(evidence, "service", service.Path, service.Platform+":"+service.Name)
		serviceAdapter := platform.NewAdapter(platform.Host{OS: service.Platform})
		if adapter != nil && service.Platform == discovery.Platform {
			serviceAdapter = adapter
		}
		switch service.Platform {
		case "darwin":
			actions = append(actions, makeActionFromEvidence(ev, model.Action{
				Kind:     model.ActionRunCommand,
				Target:   service.Name,
				Reason:   "Unload " + facts.DisplayName + " launchd service",
				Command:  serviceAdapter.ServiceDisableCommand(service),
				Platform: service.Platform,
			}))
		case "linux":
			actions = append(actions, makeActionFromEvidence(ev, model.Action{
				Kind:     model.ActionRunCommand,
				Target:   service.Name,
				Reason:   "Disable " + facts.DisplayName + " systemd service",
				Command:  serviceAdapter.ServiceDisableCommand(service),
				Platform: service.Platform,
			}))
		case "windows":
			actions = append(actions, makeActionFromEvidence(ev, model.Action{
				Kind:     model.ActionRunCommand,
				Target:   service.Name,
				Reason:   "Delete " + facts.DisplayName + " scheduled task",
				Command:  serviceAdapter.ServiceDisableCommand(service),
				Platform: service.Platform,
			}))
		}
		if service.Path != "" {
			actions = append(actions, makeActionFromEvidence(ev, model.Action{
				Kind:     model.ActionRemovePath,
				Target:   service.Path,
				Reason:   "Remove " + facts.DisplayName + " service definition file",
				Platform: service.Platform,
			}))
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
			Command:  cmd,
			Platform: discovery.Platform,
		})
		actions[len(actions)-1] = makeActionFromEvidence(ev, actions[len(actions)-1])
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

func processActions(discovery model.Discovery, evidence model.EvidenceSet, options model.Options, adapter platform.Adapter) []model.Action {
	if !options.KillProcesses {
		return nil
	}
	var actions []model.Action
	for _, proc := range discovery.Processes {
		if proc.PID == 0 {
			continue
		}
		ev := matchEvidence(evidence, "process", proc.Command)
		actions = append(actions, makeActionFromEvidence(ev, model.Action{
			Kind:     model.ActionRunCommand,
			Target:   proc.Command,
			Reason:   "Terminate matching process",
			Command:  adapter.ProcessTerminateCommand(proc.PID),
			HighRisk: true,
			Risk:     "Kills live process",
		}))
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
				HighRisk: true,
				Risk:     "Deletes container state",
			})
			actions[len(actions)-1] = makeActionFromEvidence(ev, actions[len(actions)-1])
			continue
		}
		actions = append(actions, makeActionFromEvidence(ev, model.Action{
			Kind:     model.ActionRunCommand,
			Target:   container.Runtime + ":" + container.ID,
			Reason:   "Remove matching container",
			Command:  []string{container.Runtime, "rm", "-f", container.ID},
			HighRisk: true,
			Risk:     "Deletes container state",
		}))
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
				HighRisk: true,
				Risk:     "Deletes local image",
			})
			actions[len(actions)-1] = makeActionFromEvidence(ev, actions[len(actions)-1])
			continue
		}
		target := image.ID
		if target == "" {
			target = image.Name
		}
		actions = append(actions, makeActionFromEvidence(ev, model.Action{
			Kind:     model.ActionRunCommand,
			Target:   image.Runtime + ":" + target,
			Reason:   "Remove matching image",
			Command:  []string{image.Runtime, "rmi", "-f", target},
			HighRisk: true,
			Risk:     "Deletes local image",
		}))
	}
	return actions
}

func pathActions(discovery model.Discovery, evidence model.EvidenceSet, facts model.ProductFacts, options model.Options) []model.Action {
	var out []model.Action
	appendRemove := func(kind string, paths []string, reason string) {
		for _, item := range paths {
			ev := matchEvidence(evidence, kind, item)
			out = append(out, model.Action{
				Kind:   model.ActionRemovePath,
				Target: item,
				Reason: reason,
			})
			out[len(out)-1] = makeActionFromEvidence(ev, out[len(out)-1])
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
			Kind:       model.ActionReportOnly,
			Target:     item.Target,
			Reason:     item.Reason,
			Evidence:   item.Strength,
			Rule:       item.Rule,
			Source:     item.Source,
			Confidence: item.Confidence,
			HighRisk:   item.Risk == "high",
			Risk:       item.Risk,
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

func makeActionFromEvidence(ev model.Evidence, action model.Action) model.Action {
	action.Evidence = ev.Strength
	action.Rule = ev.Rule
	action.Source = ev.Source
	action.Confidence = ev.Confidence
	if action.Kind == model.ActionReportOnly {
		return action
	}
	if action.Confidence > 0 && action.Confidence < minConfidenceForDestructive {
		action.Kind = model.ActionReportOnly
		action.Command = nil
		action.Reason = action.Reason + " (deferred: low-confidence evidence, manual review required)"
	}
	return action
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
