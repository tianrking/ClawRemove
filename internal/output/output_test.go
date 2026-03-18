package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
)

func TestPrintProductsJSON(t *testing.T) {
	facts := []model.ProductFacts{
		{ID: "openclaw", DisplayName: "OpenClaw"},
		{ID: "legacyclaw", DisplayName: "LegacyClaw"},
	}

	var buf bytes.Buffer
	err := PrintProducts(&buf, facts, FormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputStr := buf.String()
	var parsed []map[string]string
	if err := json.Unmarshal([]byte(outputStr), &parsed); err != nil {
		t.Fatalf("failed to parse JSON product list: %v", err)
	}

	if len(parsed) != 2 {
		t.Fatalf("expected 2 products, got %d", len(parsed))
	}

	if parsed[0]["id"] != "openclaw" || parsed[0]["displayName"] != "OpenClaw" {
		t.Errorf("product 0 mismatch: %v", parsed[0])
	}
}

func TestPrintReportJSON(t *testing.T) {
	report := model.Report{
		OK:      true,
		Product: "mock",
		Command: "audit",
		Host: model.Host{
			OS:   "linux",
			Arch: "amd64",
		},
		Verify: model.Verification{
			Verified: true,
			Confirmed: []model.Residual{
				{Kind: "state_dir", Target: "/tmp/mock", Confidence: 1.0, Evidence: "exact"},
			},
		},
	}

	var buf bytes.Buffer
	err := PrintReport(&buf, report, FormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputStr := buf.String()

	// Ensure the root schema contains expected elements matching the contract
	if !strings.Contains(outputStr, `"ok": true`) {
		t.Error("missing standard 'ok: true' in JSON string")
	}
	if !strings.Contains(outputStr, `"state_dir"`) {
		t.Error("missing confirmed residual struct kind in JSON string")
	}

	var roundtrip model.Report
	if err := json.Unmarshal([]byte(outputStr), &roundtrip); err != nil {
		t.Fatalf("failed to decode emitted JSON report format: %v", err)
	}

	if roundtrip.Product != "mock" {
		t.Errorf("expected product 'mock', got %s", roundtrip.Product)
	}
}
