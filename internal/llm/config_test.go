package llm

import "testing"

func TestDefaultBaseURLAndModel(t *testing.T) {
	if defaultBaseURL("openai") != "https://api.openai.com/v1" {
		t.Fatal("unexpected openai base url")
	}
	if defaultBaseURL("anthropic") != "https://api.anthropic.com/v1" {
		t.Fatal("unexpected anthropic base url")
	}
	if defaultBaseURL("openrouter") != "https://openrouter.ai/api/v1" {
		t.Fatal("unexpected openrouter base url")
	}
	if defaultBaseURL("zhipu") != "https://open.bigmodel.cn/api/paas/v4" {
		t.Fatal("unexpected zhipu base url")
	}
	if defaultModel("anthropic") == "" {
		t.Fatal("expected anthropic default model")
	}
}

func TestLoadConfigFromEnvSupportsMultipleDrivers(t *testing.T) {
	t.Setenv("CLAWREMOVE_LLM_PROVIDERS", "openai,openrouter,zhipu")
	t.Setenv("OPENAI_API_KEY", "openai-key")
	t.Setenv("OPENROUTER_API_KEY", "openrouter-key")
	t.Setenv("ZHIPU_API_KEY", "zhipu-key")
	t.Setenv("CLAWREMOVE_LLM_MODELS", "model-a,model-b")
	cfg := LoadConfigFromEnv()
	if !cfg.Enabled {
		t.Fatal("expected config to be enabled")
	}
	if len(cfg.Drivers) != 3 {
		t.Fatalf("expected 3 drivers, got %d", len(cfg.Drivers))
	}
	if cfg.Drivers[0].Provider != "openai" {
		t.Fatalf("unexpected first provider: %s", cfg.Drivers[0].Provider)
	}
	if len(cfg.Drivers[0].Models) != 2 {
		t.Fatalf("expected shared model list, got %d", len(cfg.Drivers[0].Models))
	}
}
