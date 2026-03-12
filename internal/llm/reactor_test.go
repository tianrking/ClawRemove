package llm

import (
	"context"
	"testing"

	"github.com/tianrking/ClawRemove/internal/llm/mediation"
	llmproviders "github.com/tianrking/ClawRemove/internal/llm/providers"
	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/platform"
	"github.com/tianrking/ClawRemove/internal/system"
)

type fakeClient struct {
	responses []string
	index     int
}

func (f *fakeClient) CompleteJSON(_ context.Context, _ string, _ []llmproviders.Message) (string, error) {
	response := f.responses[f.index]
	f.index++
	return response, nil
}

func TestControlledAdvisorRunsReadOnlyToolLoop(t *testing.T) {
	client := &fakeClient{
		responses: []string{
			`{"kind":"tool","thoughtSummary":"Need services","tool":"services","input":{"limit":5}}`,
			`{"kind":"final","thoughtSummary":"Review services before path cleanup.","userMessage":"Unload persistent services first.","riskNotes":["Services may restart removed binaries."],"recommendations":[{"kind":"review_services","target":"1 matching service","reason":"Persistent service found.","risk":"medium","optIn":false,"evidence":"strong"}]}`,
		},
	}
	advisor := controlledAdvisor{
		client: client,
		config: Config{Enabled: true, MaxSteps: 4},
	}
	report := model.Report{
		Product: "openclaw",
		Host:    model.Host{OS: "darwin", Arch: "arm64"},
		Discovery: model.Discovery{
			Services: []model.ServiceRef{{Platform: "darwin", Scope: "user", Name: "ai.openclaw.gateway"}},
		},
	}

	advice := advisor.Assess(context.Background(), report)
	if advice.Mode != "react-controlled" {
		t.Fatalf("expected react-controlled mode, got %q", advice.Mode)
	}
	if len(advice.Recommendations) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(advice.Recommendations))
	}
	if advice.Recommendations[0].Kind != "review_services" {
		t.Fatalf("unexpected recommendation kind: %q", advice.Recommendations[0].Kind)
	}
}

func TestExecuteToolRejectsUnsupportedTool(t *testing.T) {
	mediator := mediation.New(system.NewRunner(), platform.NewAdapter(platform.Host{OS: "darwin"}))
	_, err := mediator.ExecuteTool(model.Report{}, "destroy_system", map[string]any{})
	if err == nil {
		t.Fatal("expected unsupported tool error")
	}
}

func TestPathProbeRejectsUnknownTarget(t *testing.T) {
	mediator := mediation.New(system.NewRunner(), platform.NewAdapter(platform.Host{OS: "darwin"}))
	report := model.Report{
		Discovery: model.Discovery{
			StateDirs: []string{"/tmp/.openclaw"},
		},
	}
	_, err := mediator.ExecuteTool(report, "path_probe", map[string]any{"target": "/tmp/not-allowed"})
	if err == nil {
		t.Fatal("expected path probe to reject unknown target")
	}
}
