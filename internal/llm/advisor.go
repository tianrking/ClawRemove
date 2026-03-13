package llm

import (
	"context"
	"fmt"
	"io"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/skills"
)

// StreamFunc is a callback for streaming AI progress output.
type StreamFunc func(format string, args ...any)

// NilStreamFunc is a no-op streaming function.
func NilStreamFunc(format string, args ...any) {}

// WriterStreamFunc returns a StreamFunc that writes to an io.Writer.
func WriterStreamFunc(w io.Writer) StreamFunc {
	return func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
}

type Advisor interface {
	Assess(context.Context, model.Report, []skills.Skill) model.Advice
	AssessWithStream(context.Context, model.Report, []skills.Skill, StreamFunc) model.Advice
}

type NoopAdvisor struct{}

func NewNoopAdvisor() NoopAdvisor {
	return NoopAdvisor{}
}

func (NoopAdvisor) Assess(_ context.Context, report model.Report, _ []skills.Skill) model.Advice {
	return NoopAdvisor{}.AssessWithStream(nil, report, nil, NilStreamFunc)
}

func (NoopAdvisor) AssessWithStream(_ context.Context, report model.Report, _ []skills.Skill, _ StreamFunc) model.Advice {
	advice := model.Advice{
		Mode:            "controlled",
		Authority:       "advisory-only",
		ThoughtSummary:  "Deterministic engine review completed. Advisory output is informational and does not approve destructive execution on its own.",
		RiskNotes:       []string{"High-risk actions still require explicit operator opt-in."},
		UserMessage:     "Review the generated plan before applying removal actions.",
		NeededEvidence:  []string{},
		Recommendations: []model.Recommendation{},
	}

	if report.Verify.Verified {
		advice.ThoughtSummary = fmt.Sprintf(
			"Verification classified residuals into exact=%d, strong=%d, heuristic=%d. Confirmed residue should be prioritized before heuristic investigation.",
			report.Verify.Summary.Exact,
			report.Verify.Summary.Strong,
			report.Verify.Summary.Heuristic,
		)
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
	if len(report.Verify.Confirmed) > 0 {
		advice.Recommendations = append(advice.Recommendations, model.Recommendation{
			Kind:     "remove_confirmed_residue",
			Target:   fmt.Sprintf("%d confirmed residual(s)", len(report.Verify.Confirmed)),
			Reason:   "Exact and strong evidence indicate these leftovers still belong to the target product.",
			Risk:     "medium",
			OptIn:    false,
			Evidence: "strong",
		})
	}
	if len(report.Verify.Investigate) > 0 {
		advice.Recommendations = append(advice.Recommendations, model.Recommendation{
			Kind:     "investigate_heuristics",
			Target:   fmt.Sprintf("%d heuristic residual(s)", len(report.Verify.Investigate)),
			Reason:   "These findings should be reviewed before turning them into deletion actions.",
			Risk:     "low",
			OptIn:    false,
			Evidence: "heuristic",
		})
	}
	if report.Command == "explain" {
		advice.UserMessage = "This explanation summarizes what ClawRemove found and what should be reviewed before removal."
	}

	return advice
}
