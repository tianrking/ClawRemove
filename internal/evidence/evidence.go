package evidence

import "github.com/tianrking/ClawRemove/internal/model"

func Build(discovery model.Discovery, facts model.ProductFacts) model.EvidenceSet {
	var items []model.Evidence

	appendItems := func(kind string, targets []string, strength, reason, risk string) {
		for _, target := range targets {
			items = append(items, model.Evidence{
				Kind:     kind,
				Target:   target,
				Strength: strength,
				Reason:   reason,
				Risk:     risk,
			})
		}
	}

	appendItems("state_dir", discovery.StateDirs, "exact", facts.DisplayName+" state directory still exists", "medium")
	appendItems("workspace_dir", discovery.WorkspaceDirs, "exact", facts.DisplayName+" workspace still exists", "medium")
	appendItems("temp_path", discovery.TempPaths, "strong", facts.DisplayName+" temp or install staging path still exists", "low")
	appendItems("app_path", discovery.AppPaths, "exact", facts.DisplayName+" app artifact still exists", "medium")
	appendItems("cli_path", discovery.CLIPaths, "strong", facts.DisplayName+" CLI wrapper still exists", "low")
	appendItems("shell_profile", discovery.ShellProfiles, "heuristic", "Shell profile may still contain claw completion or bootstrap lines", "low")
	appendItems("listener", discovery.Listeners, "heuristic", "Listener still matches claw markers or default ports", "high")
	appendItems("crontab", discovery.CrontabLines, "heuristic", "Crontab line references claw markers", "high")

	for _, pkg := range discovery.Packages {
		items = append(items, model.Evidence{
			Kind:     "package",
			Target:   pkg.Manager + ":" + pkg.Name,
			Strength: "strong",
			Reason:   facts.DisplayName + " package is still installed",
			Risk:     "medium",
		})
	}
	for _, svc := range discovery.Services {
		target := svc.Platform + ":" + svc.Name
		if svc.Path != "" {
			target = svc.Path
		}
		items = append(items, model.Evidence{
			Kind:     "service",
			Target:   target,
			Strength: "strong",
			Reason:   facts.DisplayName + " service registration still exists",
			Risk:     "medium",
		})
	}
	for _, proc := range discovery.Processes {
		items = append(items, model.Evidence{
			Kind:     "process",
			Target:   proc.Command,
			Strength: "strong",
			Reason:   facts.DisplayName + " live process is still running",
			Risk:     "high",
		})
	}
	for _, c := range discovery.Containers {
		items = append(items, model.Evidence{
			Kind:     "container",
			Target:   c.Runtime + ":" + c.ID,
			Strength: "strong",
			Reason:   "Container still references claw markers",
			Risk:     "high",
		})
	}
	for _, img := range discovery.Images {
		target := img.Name
		if img.ID != "" {
			target = img.ID
		}
		items = append(items, model.Evidence{
			Kind:     "image",
			Target:   img.Runtime + ":" + target,
			Strength: "heuristic",
			Reason:   "Image still references claw markers",
			Risk:     "medium",
		})
	}

	set := model.EvidenceSet{Items: items}
	for _, item := range items {
		switch item.Strength {
		case "exact":
			set.Summary.Exact++
		case "strong":
			set.Summary.Strong++
		default:
			set.Summary.Heuristic++
		}
	}
	return set
}
