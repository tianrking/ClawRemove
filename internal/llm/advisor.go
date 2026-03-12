package llm

import (
	"context"
	"fmt"

	"github.com/tianrking/ClawRemove/internal/model"
)

type Advisor interface {
	Assess(context.Context, model.Report) model.Advice
}

type NoopAdvisor struct{}

func NewNoopAdvisor() NoopAdvisor {
	return NoopAdvisor{}
}

func (NoopAdvisor) Assess(_ context.Context, report model.Report) model.Advice {
	advice := model.Advice{
		Mode:            "controlled",
		Authority:       "advisory-only",
		ThoughtSummary:  "Deterministic engine review completed. Advisory output is informational and does not approve destructive execution on its own.",
		RiskNotes:       []string{"High-risk actions still require explicit operator opt-in."},
		UserMessage:     "Review the generated plan before applying removal actions.",
		NeededEvidence:  []string{},
		Recommendations: []model.Recommendation{},
	}

	if len(report.Discovery.Processes) > 0 {
		advice.Recommendations = append(advice.Recommendations, model.Recommendation{
			Kind:     "review_processes",
			Target:   fmt.Sprintf("%d matching process(es)", len(report.Discovery.Processes)),
			Reason:   "Live processes may hold files open or restart services during removal.",
			Risk:     "high",
			OptIn:    true,
			Evidence: "strong",
		})
	}
	if len(report.Discovery.Services) > 0 {
		advice.Recommendations = append(advice.Recommendations, model.Recommendation{
			Kind:     "review_services",
			Target:   fmt.Sprintf("%d matching service registration(s)", len(report.Discovery.Services)),
			Reason:   "Persistent services should be unloaded before path removal.",
			Risk:     "medium",
			OptIn:    false,
			Evidence: "strong",
		})
	}
	if len(report.Discovery.StateDirs)+len(report.Discovery.WorkspaceDirs)+len(report.Discovery.AppPaths) == 0 {
		advice.UserMessage = "No major filesystem residue was discovered for this provider on the current host."
	}
	if report.Command == "explain" {
		advice.UserMessage = "This explanation summarizes what ClawRemove found and what should be reviewed before removal."
	}

	return advice
}
