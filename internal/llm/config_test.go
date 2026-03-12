package llm

import "testing"

func TestDefaultBaseURLAndModel(t *testing.T) {
	if defaultBaseURL("openai") != "https://api.openai.com/v1" {
		t.Fatal("unexpected openai base url")
	}
	if defaultBaseURL("anthropic") != "https://api.anthropic.com/v1" {
		t.Fatal("unexpected anthropic base url")
	}
	if defaultModel("anthropic") == "" {
		t.Fatal("expected anthropic default model")
	}
}
