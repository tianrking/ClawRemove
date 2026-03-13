package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type DriverConfig struct {
	Provider string
	BaseURL  string
	APIKey   string
	Models   []string
}

type Config struct {
	Drivers        []DriverConfig
	MaxTokens      int
	TimeoutSeconds int
	UserAgent      string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Client interface {
	CompleteJSON(ctx context.Context, systemPrompt string, messages []Message) (string, error)
}

type Trace struct {
	Selected string
	Attempts []string
}

type TraceClient interface {
	CompleteJSONWithTrace(ctx context.Context, systemPrompt string, messages []Message) (string, Trace, error)
}

type candidate struct {
	ID     string
	Client modelClient
}

type modelClient interface {
	CompleteJSON(ctx context.Context, systemPrompt string, messages []Message) (string, error)
}

type chainClient struct {
	candidates []candidate
}

type singleConfig struct {
	Provider  string
	BaseURL   string
	APIKey    string
	Model     string
	MaxTokens int
	Timeout   time.Duration
	UserAgent string
}

type openAICompatibleClient struct {
	httpClient *http.Client
	config     singleConfig
}

type anthropicClient struct {
	httpClient *http.Client
	config     singleConfig
}

type chatCompletionRequest struct {
	Model          string    `json:"model"`
	Messages       []Message `json:"messages"`
	ResponseFormat struct {
		Type string `json:"type"`
	} `json:"response_format"`
	MaxTokens int `json:"max_tokens,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type anthropicMessageRequest struct {
	Model     string    `json:"model"`
	System    string    `json:"system"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type anthropicMessageResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func NewFromConfig(cfg Config) Client {
	timeout := secondsToDuration(cfg.TimeoutSeconds)
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 1200
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "ClawRemove/0"
	}
	var out []candidate
	for _, driver := range cfg.Drivers {
		for _, model := range driver.Models {
			sc := singleConfig{
				Provider:  strings.ToLower(driver.Provider),
				BaseURL:   strings.TrimRight(driver.BaseURL, "/"),
				APIKey:    driver.APIKey,
				Model:     model,
				MaxTokens: cfg.MaxTokens,
				Timeout:   timeout,
				UserAgent: cfg.UserAgent,
			}
			client := newSingleClient(sc)
			if client == nil {
				continue
			}
			out = append(out, candidate{
				ID:     sc.Provider + ":" + sc.Model,
				Client: client,
			})
		}
	}
	return chainClient{candidates: out}
}

func newSingleClient(cfg singleConfig) modelClient {
	switch cfg.Provider {
	case "anthropic", "anthropic-compatible":
		return anthropicClient{
			httpClient: &http.Client{Timeout: cfg.Timeout},
			config:     cfg,
		}
	case "openai", "openai-compatible", "openrouter", "zhipu":
		return openAICompatibleClient{
			httpClient: &http.Client{Timeout: cfg.Timeout},
			config:     cfg,
		}
	default:
		return nil
	}
}

func (c chainClient) CompleteJSON(ctx context.Context, systemPrompt string, messages []Message) (string, error) {
	content, _, err := c.CompleteJSONWithTrace(ctx, systemPrompt, messages)
	return content, err
}

func (c chainClient) CompleteJSONWithTrace(ctx context.Context, systemPrompt string, messages []Message) (string, Trace, error) {
	if len(c.candidates) == 0 {
		return "", Trace{}, fmt.Errorf("no llm candidates configured")
	}
	var errs []string
	var attempts []string
	for _, candidate := range c.candidates {
		content, err := candidate.Client.CompleteJSON(ctx, systemPrompt, messages)
		if err == nil {
			attempts = append(attempts, candidate.ID+":ok")
			return content, Trace{
				Selected: candidate.ID,
				Attempts: attempts,
			}, nil
		}
		errs = append(errs, candidate.ID+": "+err.Error())
		attempts = append(attempts, candidate.ID+":fail")
	}
	return "", Trace{Attempts: attempts}, fmt.Errorf("all llm candidates failed: %s", strings.Join(errs, " | "))
}

func secondsToDuration(seconds int) time.Duration {
	if seconds <= 0 {
		seconds = 45
	}
	return time.Duration(seconds) * time.Second
}

func (c openAICompatibleClient) CompleteJSON(ctx context.Context, systemPrompt string, messages []Message) (string, error) {
	payload := chatCompletionRequest{
		Model:     c.config.Model,
		Messages:  append([]Message{{Role: "system", Content: systemPrompt}}, messages...),
		MaxTokens: c.config.MaxTokens,
	}
	payload.ResponseFormat.Type = "json_object"
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal llm request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create llm request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.config.UserAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send llm request: %w", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read llm response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("llm request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	var parsed chatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return "", fmt.Errorf("decode llm response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("llm response returned no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

func (c anthropicClient) CompleteJSON(ctx context.Context, systemPrompt string, messages []Message) (string, error) {
	payload := anthropicMessageRequest{
		Model:     c.config.Model,
		System:    systemPrompt,
		Messages:  messages,
		MaxTokens: c.config.MaxTokens,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal anthropic request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.BaseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create anthropic request: %w", err)
	}
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("User-Agent", c.config.UserAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send anthropic request: %w", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read anthropic response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("anthropic request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	var parsed anthropicMessageResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return "", fmt.Errorf("decode anthropic response: %w", err)
	}
	for _, item := range parsed.Content {
		if item.Type == "text" && item.Text != "" {
			return item.Text, nil
		}
	}
	return "", fmt.Errorf("anthropic response returned no text content")
}
