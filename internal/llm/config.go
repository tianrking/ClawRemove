package llm

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Enabled   bool
	Provider  string
	BaseURL   string
	APIKey    string
	Model     string
	MaxTokens int
	MaxSteps  int
	Timeout   time.Duration
	UserAgent string
}

func LoadConfigFromEnv() Config {
	provider := strings.ToLower(envOr("CLAWREMOVE_LLM_PROVIDER", "openai"))
	cfg := Config{
		Provider:  provider,
		Model:     envOr("CLAWREMOVE_LLM_MODEL", defaultModel(provider)),
		MaxTokens: envOrInt("CLAWREMOVE_LLM_MAX_TOKENS", 1200),
		MaxSteps:  envOrInt("CLAWREMOVE_LLM_MAX_STEPS", 4),
		Timeout:   time.Duration(envOrInt("CLAWREMOVE_LLM_TIMEOUT_SECONDS", 45)) * time.Second,
		UserAgent: envOr("CLAWREMOVE_LLM_USER_AGENT", "ClawRemove/0"),
	}
	cfg.BaseURL = strings.TrimRight(envOr("CLAWREMOVE_LLM_BASE_URL", defaultBaseURL(provider)), "/")
	cfg.APIKey = resolveAPIKey(provider)
	cfg.Enabled = cfg.APIKey != "" && cfg.Model != "" && cfg.BaseURL != ""
	return cfg
}

func defaultBaseURL(provider string) string {
	switch provider {
	case "anthropic":
		return "https://api.anthropic.com/v1"
	case "openai-compatible":
		return "https://api.openai.com/v1"
	default:
		return "https://api.openai.com/v1"
	}
}

func defaultModel(provider string) string {
	switch provider {
	case "anthropic":
		return "claude-3-5-sonnet-latest"
	default:
		return "gpt-4.1-mini"
	}
}

func resolveAPIKey(provider string) string {
	if value := strings.TrimSpace(os.Getenv("CLAWREMOVE_LLM_API_KEY")); value != "" {
		return value
	}
	switch provider {
	case "anthropic":
		return strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	default:
		return strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	}
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envOrInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
