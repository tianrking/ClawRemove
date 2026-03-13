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
	return a.AssessWithStream(ctx, report, skills, NilStreamFunc)
}

func (a controlledAdvisor) AssessWithStream(ctx context.Context, report model.Report, skills []skills.Skill, stream StreamFunc) model.Advice {
	base := NewNoopAdvisor().AssessWithStream(ctx, report, skills, stream)
	base.Mode = "react-controlled"
	base.UserMessage = "Controlled advisor is enabled. Review its recommendations as guidance, not as execution authority."

	stream("🤖 AI Analysis Starting...")
	stream("   Provider: %s", report.Product)
	stream("   Command: %s", report.Command)

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
		stream("", "")
		stream("🔄 ReAct Step %d/%d...", step+1, a.config.MaxSteps)

		var (
			content string
			err     error
		)
		if traceClient, ok := a.client.(llmproviders.TraceClient); ok {
			var trace llmproviders.Trace
			stream("   📤 Calling LLM...", "")
			content, trace, err = traceClient.CompleteJSONWithTrace(ctx, systemPrompt, messages)
			if a.config.Trace && len(trace.Attempts) > 0 {
				base.Trace = append(base.Trace, "llm-attempts["+itoa(step)+"]="+strings.Join(trace.Attempts, " -> "))
			}
			if a.config.Trace && trace.Selected != "" {
				base.Trace = append(base.Trace, "llm-selected["+itoa(step)+"]="+trace.Selected)
			}
		} else {
			stream("   📤 Calling LLM...", "")
			content, err = a.client.CompleteJSON(ctx, systemPrompt, messages)
		}
		if err != nil {
			stream("   ❌ LLM error: %s", err.Error())
			base.RiskNotes = append(base.RiskNotes, "LLM advisor fallback: "+err.Error())
			return base
		}

		var next reactorStep
		if err := json.Unmarshal([]byte(content), &next); err != nil {
			stream("   ❌ Invalid JSON response", "")
			base.RiskNotes = append(base.RiskNotes, "LLM advisor returned invalid JSON; deterministic fallback was used.")
			return base
		}

		if next.ThoughtSummary != "" {
			stream("   💭 Thought: %s", truncateText(next.ThoughtSummary, 100))
		}

		switch strings.ToLower(next.Kind) {
		case "tool":
			stream("   🔧 Using tool: %s", next.Tool)
			toolResult, toolErr := a.mediator.ExecuteTool(ctx, report, next.Tool, next.Input)
			if toolErr != nil {
				stream("   ❌ Tool error: %s", toolErr.Error())
				base.RiskNotes = append(base.RiskNotes, "LLM requested invalid tool input; deterministic fallback was used.")
				return base
			}
			stream("   ✅ Tool result received", "")
			messages = append(messages,
				llmproviders.Message{Role: "assistant", Content: content},
				llmproviders.Message{Role: "user", Content: fmt.Sprintf("Tool result for %s:\n%s", next.Tool, mustJSON(toolResult))},
			)
		case "final":
			stream("", "")
			stream("✅ AI Analysis Complete!", "")
			if next.UserMessage != "" {
				stream("   📝 Summary: %s", truncateText(next.UserMessage, 150))
			}
			return mergeAdvice(base, next)
		default:
			stream("   ⚠️ Unknown response kind: %s", next.Kind)
			base.RiskNotes = append(base.RiskNotes, "LLM advisor returned an unsupported response kind; deterministic fallback was used.")
			return base
		}
	}

	stream("", "")
	stream("⚠️ Max reasoning steps reached", "")
	base.RiskNotes = append(base.RiskNotes, "LLM advisor reached the maximum number of controlled reasoning steps.")
	return base
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
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
