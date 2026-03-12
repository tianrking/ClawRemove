package plan

import (
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
)

func TestBuildKeepsWorkspaceWhenRequested(t *testing.T) {
	discovery := model.Discovery{
		StateDirs:     []string{"/tmp/.openclaw"},
		WorkspaceDirs: []string{"/tmp/.openclaw/workspace"},
	}
	facts := model.ProductFacts{DisplayName: "OpenClaw"}

	plan := Build(discovery, facts, model.Options{KeepWorkspace: true})
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

	plan := Build(discovery, facts, model.Options{})
	if len(plan.Actions) == 0 || plan.Actions[0].Kind != model.ActionReportOnly {
		t.Fatalf("expected report-only action for container without remove flag")
	}
}
