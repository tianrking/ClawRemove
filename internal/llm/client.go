package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type chatClient interface {
	CompleteJSON(ctx context.Context, systemPrompt string, messages []chatMessage) (string, error)
}

type openAICompatibleClient struct {
	httpClient *http.Client
	config     Config
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	ResponseFormat responseFormat `json:"response_format"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func newOpenAICompatibleClient(cfg Config) chatClient {
	return openAICompatibleClient{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		config:     cfg,
	}
}

func (c openAICompatibleClient) CompleteJSON(ctx context.Context, systemPrompt string, messages []chatMessage) (string, error) {
	payload := chatCompletionRequest{
		Model:          c.config.Model,
		ResponseFormat: responseFormat{Type: "json_object"},
		Messages:       append([]chatMessage{{Role: "system", Content: systemPrompt}}, messages...),
	}
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

	var parsed chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decode llm response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("llm request failed with status %d", resp.StatusCode)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("llm response returned no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}
