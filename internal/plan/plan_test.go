package plan

import (
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/platform"
)

func TestBuildKeepsWorkspaceWhenRequested(t *testing.T) {
	discovery := model.Discovery{
		StateDirs:     []string{"/tmp/.openclaw"},
		WorkspaceDirs: []string{"/tmp/.openclaw/workspace"},
	}
	facts := model.ProductFacts{DisplayName: "OpenClaw"}
	evidence := model.EvidenceSet{}

	plan := Build(discovery, evidence, facts, model.Options{KeepWorkspace: true}, platform.Host{OS: "linux"})
	for _, action := range plan.Actions {
		if action.Target == "/tmp/.openclaw/workspace" {
			t.Fatalf("workspace removal should not be planned when KeepWorkspace is true")
		}
	}
}

func TestBuildReportsContainerWithoutRemovalFlag(t *testing.T) {
	discovery := model.Discovery{
		Containers: []model.ContainerRef{{Runtime: "docker", ID: "abc"}},
	}
	facts := model.ProductFacts{DisplayName: "OpenClaw"}
	evidence := model.EvidenceSet{
		Items: []model.Evidence{{Kind: "container", Target: "docker:abc", Strength: "strong"}},
	}

	plan := Build(discovery, evidence, facts, model.Options{}, platform.Host{OS: "linux"})
	if len(plan.Actions) == 0 || plan.Actions[0].Kind != model.ActionReportOnly {
		t.Fatalf("expected report-only action for container without remove flag")
	}
}

func TestBuildDefersLowConfidenceDestructiveAction(t *testing.T) {
	discovery := model.Discovery{
		TempPaths: []string{"/tmp/openclaw-123"},
	}
	facts := model.ProductFacts{DisplayName: "OpenClaw"}
	evidence := model.EvidenceSet{
		Items: []model.Evidence{{
			Kind:       "temp_path",
			Target:     "/tmp/openclaw-123",
			Strength:   "heuristic",
			Rule:       "temp-path-marker",
			Source:     "filesystem",
			Confidence: 0.55,
		}},
	}

	plan := Build(discovery, evidence, facts, model.Options{}, platform.Host{OS: "linux"})
	if len(plan.Actions) == 0 {
		t.Fatalf("expected at least one action")
	}
	if plan.Actions[0].Kind != model.ActionReportOnly {
		t.Fatalf("expected low-confidence path action to be report-only, got %s", plan.Actions[0].Kind)
	}
	if plan.Actions[0].Rule != "temp-path-marker" || plan.Actions[0].Source != "filesystem" {
		t.Fatalf("expected provenance to be attached to action")
	}
}
