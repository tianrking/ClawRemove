package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Config struct {
	Provider       string
	BaseURL        string
	APIKey         string
	Model          string
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

type openAICompatibleClient struct {
	httpClient *http.Client
	config     Config
}

type anthropicClient struct {
	httpClient *http.Client
	config     Config
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
	switch cfg.Provider {
	case "anthropic":
		return anthropicClient{
			httpClient: &http.Client{Timeout: secondsToDuration(cfg.TimeoutSeconds)},
			config:     cfg,
		}
	default:
		return openAICompatibleClient{
			httpClient: &http.Client{Timeout: secondsToDuration(cfg.TimeoutSeconds)},
			config:     cfg,
		}
	}
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
