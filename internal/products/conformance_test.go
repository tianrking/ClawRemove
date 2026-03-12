package products_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/products"
)

func TestProviderConformance(t *testing.T) {
	registry := products.Registry()
	if len(registry) == 0 {
		t.Fatal("registry is empty, expected at least one provider (e.g., openclaw)")
	}

	idPattern := regexp.MustCompile(`^[a-z0-9-]+$`)

	for _, p := range registry {
		t.Run(p.ID(), func(t *testing.T) {
			// 1. Basic Metadata Conformance
			if !idPattern.MatchString(p.ID()) {
				t.Errorf("provider ID %q contains invalid characters (must be lowercase alphanumeric and hyphens)", p.ID())
			}

			if p.DisplayName() == "" {
				t.Error("provider DisplayName is empty")
			}

			// 2. Capability Parity Conformance
			caps := p.Capabilities()
			skills := p.Skills()
			tools := p.Tools()

			if len(caps.Skills) != len(skills) {
				t.Errorf("Capabilities() reported %d skills, but Skills() returned %d contracts", len(caps.Skills), len(skills))
			}

			if len(caps.Tools) != len(tools) {
				t.Errorf("Capabilities() reported %d tools, but Tools() returned %d contracts", len(caps.Tools), len(tools))
			}

			// 3. Tool Contract Conformance
			for _, tool := range tools {
				info := tool.Info()
				if info.ID == "" {
					t.Errorf("tool is missing ID: %v", info)
				}
				if info.Description == "" {
					t.Errorf("tool %q is missing a description", info.ID)
				}
				
				// Safety check: tool should not panic on empty/nil inputs
				_, err := tool.Execute(context.Background(), model.Report{}, map[string]any{})
				if err == nil {
					// Fallbacks and panics are what we care about. Returning nil/nil is technically allowed 
					// for dry-runs depending on structural definition (e.g., generic fallbacks).
					// If missing arguments occur, we usually expect an error, but as a generic conformance suite,
					// we just enforce that it doesn't crash the host test runner.
				}
			}

			// 4. Skill Contract Conformance
			for _, skill := range skills {
				info := skill.Info()
				if info.ID == "" {
					t.Errorf("skill is missing ID: %v", info)
				}
				if info.Description == "" {
					t.Errorf("skill %q is missing a description", info.ID)
				}

				// Safety check: skill should not panic on nil inputs
				_, _ = skill.Analyze(context.Background(), model.Report{})
			}
		})
	}
}
