package verify

import (
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
)

func TestClassifySplitsConfirmedAndInvestigate(t *testing.T) {
	discovery := model.Discovery{
		StateDirs:     []string{"/tmp/.openclaw"},
		Packages:      []model.PackageRef{{Manager: "npm", Name: "openclaw"}},
		Listeners:     []string{"tcp LISTEN 127.0.0.1:3456"},
		ShellProfiles: []string{"/tmp/.zshrc"},
	}
	facts := model.ProductFacts{DisplayName: "OpenClaw"}

	verification := Classify(discovery, facts)
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
