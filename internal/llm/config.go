package llm

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Enabled   bool
	Drivers   []Driver
	MaxTokens int
	MaxSteps  int
	Trace     bool
	Timeout   time.Duration
	UserAgent string
}

type Driver struct {
	Provider string
	BaseURL  string
	APIKey   string
	Models   []string
}

func LoadConfigFromEnv() Config {
	providers := parseProviders(envOr("CLAWREMOVE_LLM_PROVIDERS", envOr("CLAWREMOVE_LLM_PROVIDER", "openai")))
	cfg := Config{
		MaxTokens: envOrInt("CLAWREMOVE_LLM_MAX_TOKENS", 1200),
		MaxSteps:  envOrInt("CLAWREMOVE_LLM_MAX_STEPS", 20), // Allow AI to reason until complete
		Trace:     envOrBool("CLAWREMOVE_LLM_TRACE", false),
		Timeout:   time.Duration(envOrInt("CLAWREMOVE_LLM_TIMEOUT_SECONDS", 45)) * time.Second,
		UserAgent: envOr("CLAWREMOVE_LLM_USER_AGENT", "ClawRemove/0"),
	}
	for _, provider := range providers {
		driver := buildDriver(provider)
		if driver.APIKey == "" || len(driver.Models) == 0 || driver.BaseURL == "" {
			continue
		}
		cfg.Drivers = append(cfg.Drivers, driver)
	}
	cfg.Enabled = len(cfg.Drivers) > 0
	return cfg
}

func buildDriver(provider string) Driver {
	models := parseValues(envOr(driverEnvKey(provider, "MODELS"), envOr("CLAWREMOVE_LLM_MODELS", envOr("CLAWREMOVE_LLM_MODEL", defaultModel(provider)))))
	return Driver{
		Provider: provider,
		BaseURL: strings.TrimRight(
			envOr(driverEnvKey(provider, "BASE_URL"), envOr("CLAWREMOVE_LLM_BASE_URL", defaultBaseURL(provider))),
			"/",
		),
		APIKey: resolveAPIKey(provider),
		Models: models,
	}
}

func defaultBaseURL(provider string) string {
	switch provider {
	case "anthropic":
		return "https://api.anthropic.com/v1"
	case "openrouter":
		return "https://openrouter.ai/api/v1"
	case "zhipu":
		return "https://open.bigmodel.cn/api/paas/v4"
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
	case "openrouter":
		return "openai/gpt-4.1-mini"
	case "zhipu":
		return "glm-4.5-air"
	default:
		return "gpt-4.1-mini"
	}
}

func resolveAPIKey(provider string) string {
	if value := strings.TrimSpace(os.Getenv(driverEnvKey(provider, "API_KEY"))); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("CLAWREMOVE_LLM_API_KEY")); value != "" {
		return value
	}
	switch provider {
	case "anthropic":
		return strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	case "openrouter":
		return strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY"))
	case "zhipu":
		if value := strings.TrimSpace(os.Getenv("ZHIPU_API_KEY")); value != "" {
			return value
		}
		return strings.TrimSpace(os.Getenv("BIGMODEL_API_KEY"))
	default:
		return strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	}
}

func driverEnvKey(provider string, key string) string {
	normalized := strings.NewReplacer("-", "_", ".", "_", "/", "_").Replace(strings.ToUpper(provider))
	return "CLAWREMOVE_LLM_" + normalized + "_" + key
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

func envOrBool(key string, fallback bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if raw == "" {
		return fallback
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func parseProviders(raw string) []string {
	values := parseValues(raw)
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, strings.ToLower(value))
	}
	return out
}

func parseValues(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
