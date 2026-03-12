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
	MaxSteps  int
	Timeout   time.Duration
	UserAgent string
}

func LoadConfigFromEnv() Config {
	cfg := Config{
		Provider:  envOr("CLAWREMOVE_LLM_PROVIDER", "openai-compatible"),
		BaseURL:   strings.TrimRight(envOr("CLAWREMOVE_LLM_BASE_URL", "https://api.openai.com/v1"), "/"),
		APIKey:    os.Getenv("CLAWREMOVE_LLM_API_KEY"),
		Model:     envOr("CLAWREMOVE_LLM_MODEL", "gpt-4.1-mini"),
		MaxSteps:  envOrInt("CLAWREMOVE_LLM_MAX_STEPS", 4),
		Timeout:   time.Duration(envOrInt("CLAWREMOVE_LLM_TIMEOUT_SECONDS", 45)) * time.Second,
		UserAgent: envOr("CLAWREMOVE_LLM_USER_AGENT", "ClawRemove/0"),
	}
	cfg.Enabled = cfg.APIKey != "" && cfg.Model != "" && cfg.BaseURL != ""
	return cfg
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
