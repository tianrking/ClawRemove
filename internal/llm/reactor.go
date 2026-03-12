package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tianrking/ClawRemove/internal/llm/mediation"
	"github.com/tianrking/ClawRemove/internal/llm/prompts"
	llmproviders "github.com/tianrking/ClawRemove/internal/llm/providers"
	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/platform"
	"github.com/tianrking/ClawRemove/internal/skills"
	"github.com/tianrking/ClawRemove/internal/system"
	"github.com/tianrking/ClawRemove/internal/tools"
)

type controlledAdvisor struct {
	client   llmproviders.Client
	config   Config
	mediator mediation.Mediator
}

type reactorStep struct {
	Kind            string                 `json:"kind"`
	ThoughtSummary  string                 `json:"thoughtSummary,omitempty"`
	Tool            string                 `json:"tool,omitempty"`
	Input           map[string]any         `json:"input,omitempty"`
	NeededEvidence  []string               `json:"neededEvidence,omitempty"`
	Recommendations []model.Recommendation `json:"recommendations,omitempty"`
	RiskNotes       []string               `json:"riskNotes,omitempty"`
	UserMessage     string                 `json:"userMessage,omitempty"`
}

func NewAdvisorFromEnv(runner system.Runner, host platform.Host, providerTools []tools.Tool) Advisor {
	cfg := LoadConfigFromEnv()
	if !cfg.Enabled {
		return NewNoopAdvisor()
	}
	return controlledAdvisor{
		client: llmproviders.NewFromConfig(llmproviders.Config{
			Drivers:        toProviderDrivers(cfg.Drivers),
			MaxTokens:      cfg.MaxTokens,
			TimeoutSeconds: int(cfg.Timeout.Seconds()),
			UserAgent:      cfg.UserAgent,
		}),
		config:   cfg,
		mediator: mediation.New(runner, platform.NewAdapter(host), providerTools),
	}
}

func toProviderDrivers(drivers []Driver) []llmproviders.DriverConfig {
	out := make([]llmproviders.DriverConfig, 0, len(drivers))
	for _, driver := range drivers {
		out = append(out, llmproviders.DriverConfig{
			Provider: driver.Provider,
			BaseURL:  driver.BaseURL,
			APIKey:   driver.APIKey,
			Models:   driver.Models,
		})
	}
	return out
}

func (a controlledAdvisor) Assess(ctx context.Context, report model.Report, skills []skills.Skill) model.Advice {
	base := NewNoopAdvisor().Assess(ctx, report, skills)
	base.Mode = "react-controlled"
	base.UserMessage = "Controlled advisor is enabled. Review its recommendations as guidance, not as execution authority."

	systemPrompt := prompts.ControlledSystemPrompt()
	messages := []llmproviders.Message{
		{
			Role: "user",
			Content: fmt.Sprintf(
				"Analyze this ClawRemove report. You may request only read-only tools over the in-memory report.\nReport JSON:\n%s",
				mustJSON(reportContext(report)),
			),
		},
	}

	for step := 0; step < a.config.MaxSteps; step++ {
		var (
			content string
			err     error
		)
		if traceClient, ok := a.client.(llmproviders.TraceClient); ok {
			var trace llmproviders.Trace
			content, trace, err = traceClient.CompleteJSONWithTrace(ctx, systemPrompt, messages)
			if a.config.Trace && len(trace.Attempts) > 0 {
				base.Trace = append(base.Trace, "llm-attempts["+itoa(step)+"]="+strings.Join(trace.Attempts, " -> "))
			}
			if a.config.Trace && trace.Selected != "" {
				base.Trace = append(base.Trace, "llm-selected["+itoa(step)+"]="+trace.Selected)
			}
		} else {
			content, err = a.client.CompleteJSON(ctx, systemPrompt, messages)
		}
		if err != nil {
			base.RiskNotes = append(base.RiskNotes, "LLM advisor fallback: "+err.Error())
			return base
		}

		var next reactorStep
		if err := json.Unmarshal([]byte(content), &next); err != nil {
			base.RiskNotes = append(base.RiskNotes, "LLM advisor returned invalid JSON; deterministic fallback was used.")
			return base
		}

		switch strings.ToLower(next.Kind) {
		case "tool":
			toolResult, toolErr := a.mediator.ExecuteTool(ctx, report, next.Tool, next.Input)
			if toolErr != nil {
				base.RiskNotes = append(base.RiskNotes, "LLM requested invalid tool input; deterministic fallback was used.")
				return base
			}
			messages = append(messages,
				llmproviders.Message{Role: "assistant", Content: content},
				llmproviders.Message{Role: "user", Content: fmt.Sprintf("Tool result for %s:\n%s", next.Tool, mustJSON(toolResult))},
			)
		case "final":
			return mergeAdvice(base, next)
		default:
			base.RiskNotes = append(base.RiskNotes, "LLM advisor returned an unsupported response kind; deterministic fallback was used.")
			return base
		}
	}

	base.RiskNotes = append(base.RiskNotes, "LLM advisor reached the maximum number of controlled reasoning steps.")
	return base
}

func reportContext(report model.Report) map[string]any {
	return map[string]any{
		"product": report.Product,
		"command": report.Command,
		"host":    report.Host,
		"skills":  report.Capabilities.Skills,
		"tools":   report.Capabilities.Tools,
		"counts": map[string]int{
			"stateDirs":     len(report.Discovery.StateDirs),
			"workspaceDirs": len(report.Discovery.WorkspaceDirs),
			"services":      len(report.Discovery.Services),
			"packages":      len(report.Discovery.Packages),
			"processes":     len(report.Discovery.Processes),
			"containers":    len(report.Discovery.Containers),
			"images":        len(report.Discovery.Images),
			"planActions":   len(report.Plan.Actions),
			"evidenceItems": len(report.Evidence.Items),
			"confirmed":     len(report.Verify.Confirmed),
			"investigate":   len(report.Verify.Investigate),
		},
	}
}

func mergeAdvice(base model.Advice, next reactorStep) model.Advice {
	if next.ThoughtSummary != "" {
		base.ThoughtSummary = next.ThoughtSummary
	}
	if next.UserMessage != "" {
		base.UserMessage = next.UserMessage
	}
	if len(next.NeededEvidence) > 0 {
		base.NeededEvidence = next.NeededEvidence
	}
	if len(next.RiskNotes) > 0 {
		base.RiskNotes = append(base.RiskNotes, next.RiskNotes...)
	}
	if len(next.Recommendations) > 0 {
		base.Recommendations = next.Recommendations
	}
	return base
}

func mustJSON(value any) string {
	body, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(body)
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
