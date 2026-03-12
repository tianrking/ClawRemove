package verify

import (
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
)

func TestClassifySplitsConfirmedAndInvestigate(t *testing.T) {
	evidence := model.EvidenceSet{
		Items: []model.Evidence{
			{Kind: "state_dir", Target: "/tmp/.openclaw", Strength: "exact"},
			{Kind: "package", Target: "npm:openclaw", Strength: "strong"},
			{Kind: "listener", Target: "tcp LISTEN 127.0.0.1:3456", Strength: "heuristic"},
			{Kind: "shell_profile", Target: "/tmp/.zshrc", Strength: "heuristic"},
		},
		Summary: model.EvidenceSummary{Exact: 1, Strong: 1, Heuristic: 2},
	}

	verification := Classify(evidence)
	if !verification.Verified {
		t.Fatal("expected verification to be marked verified")
	}
	if verification.Summary.Exact != 1 {
		t.Fatalf("expected 1 exact residual, got %d", verification.Summary.Exact)
	}
	if verification.Summary.Strong != 1 {
		t.Fatalf("expected 1 strong residual, got %d", verification.Summary.Strong)
	}
	if verification.Summary.Heuristic != 2 {
		t.Fatalf("expected 2 heuristic residuals, got %d", verification.Summary.Heuristic)
	}
	if len(verification.Confirmed) != 2 {
		t.Fatalf("expected 2 confirmed residuals, got %d", len(verification.Confirmed))
	}
	if len(verification.Investigate) != 2 {
		t.Fatalf("expected 2 investigate residuals, got %d", len(verification.Investigate))
	}
}
