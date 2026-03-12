package llm

import (
	"context"
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
)

type fakeClient struct {
	responses []string
	index     int
}

func (f *fakeClient) CompleteJSON(_ context.Context, _ string, _ []chatMessage) (string, error) {
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
	_, err := executeTool(model.Report{}, "destroy_system", nil)
	if err == nil {
		t.Fatal("expected unsupported tool error")
	}
}
