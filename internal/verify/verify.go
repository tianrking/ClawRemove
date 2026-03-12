package verify

import "github.com/tianrking/ClawRemove/internal/model"

func Classify(discovery model.Discovery, facts model.ProductFacts) model.Verification {
	var all []model.Residual

	appendResiduals := func(kind string, items []string, evidence, reason, risk string) {
		for _, item := range items {
			all = append(all, model.Residual{
				Kind:     kind,
				Target:   item,
				Evidence: evidence,
				Reason:   reason,
				Risk:     risk,
			})
		}
	}

	appendResiduals("state_dir", discovery.StateDirs, "exact", facts.DisplayName+" state directory still exists", "medium")
	appendResiduals("workspace_dir", discovery.WorkspaceDirs, "exact", facts.DisplayName+" workspace still exists", "medium")
	appendResiduals("temp_path", discovery.TempPaths, "strong", facts.DisplayName+" temp or install staging path still exists", "low")
	appendResiduals("app_path", discovery.AppPaths, "exact", facts.DisplayName+" app artifact still exists", "medium")
	appendResiduals("cli_path", discovery.CLIPaths, "strong", facts.DisplayName+" CLI wrapper still exists", "low")

	for _, pkg := range discovery.Packages {
		all = append(all, model.Residual{
			Kind:     "package",
			Target:   pkg.Manager + ":" + pkg.Name,
			Evidence: "strong",
			Reason:   facts.DisplayName + " package is still installed",
			Risk:     "medium",
		})
	}
	for _, svc := range discovery.Services {
		target := svc.Platform + ":" + svc.Name
		if svc.Path != "" {
			target = svc.Path
		}
		all = append(all, model.Residual{
			Kind:     "service",
			Target:   target,
			Evidence: "strong",
			Reason:   facts.DisplayName + " service registration still exists",
			Risk:     "medium",
		})
	}
	for _, proc := range discovery.Processes {
		all = append(all, model.Residual{
			Kind:     "process",
			Target:   proc.Command,
			Evidence: "strong",
			Reason:   facts.DisplayName + " live process is still running",
			Risk:     "high",
		})
	}
	for _, item := range discovery.Listeners {
		all = append(all, model.Residual{
			Kind:     "listener",
			Target:   item,
			Evidence: "heuristic",
			Reason:   "Listener still matches claw markers or default ports",
			Risk:     "high",
		})
	}
	for _, line := range discovery.CrontabLines {
		all = append(all, model.Residual{
			Kind:     "crontab",
			Target:   line,
			Evidence: "heuristic",
			Reason:   "Crontab line references claw markers",
			Risk:     "high",
		})
	}
	for _, c := range discovery.Containers {
		all = append(all, model.Residual{
			Kind:     "container",
			Target:   c.Runtime + ":" + c.ID,
			Evidence: "strong",
			Reason:   "Container still references claw markers",
			Risk:     "high",
		})
	}
	for _, img := range discovery.Images {
		target := img.Name
		if img.ID != "" {
			target = img.ID
		}
		all = append(all, model.Residual{
			Kind:     "image",
			Target:   img.Runtime + ":" + target,
			Evidence: "heuristic",
			Reason:   "Image still references claw markers",
			Risk:     "medium",
		})
	}
	for _, profile := range discovery.ShellProfiles {
		all = append(all, model.Residual{
			Kind:     "shell_profile",
			Target:   profile,
			Evidence: "heuristic",
			Reason:   "Shell profile may still contain claw completion or bootstrap lines",
			Risk:     "low",
		})
	}

	verification := model.Verification{
		Verified:  true,
		Residuals: all,
	}
	for _, residual := range all {
		switch residual.Evidence {
		case "exact":
			verification.Summary.Exact++
			verification.Confirmed = append(verification.Confirmed, residual)
		case "strong":
			verification.Summary.Strong++
			verification.Confirmed = append(verification.Confirmed, residual)
		default:
			verification.Summary.Heuristic++
			verification.Investigate = append(verification.Investigate, residual)
		}
	}
	return verification
}
