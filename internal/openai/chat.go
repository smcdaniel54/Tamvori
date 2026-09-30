package openai

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

const defaultEndpoint = "https://api.openai.com/v1/chat/completions"

// ChatJSON calls OpenAI chat completions and returns the assistant message content
// plus the raw HTTP response body (for hashing/provenance). apiKey must never be logged.
func ChatJSON(ctx context.Context, apiKey, model, system, user string, maxTokens int) (content string, raw []byte, err error) {
	if apiKey == "" {
		return "", nil, fmt.Errorf("openai: missing api key")
	}
	if model == "" {
		return "", nil, fmt.Errorf("openai: missing model")
	}
	if maxTokens <= 0 {
		maxTokens = 800
	}
	body := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature":     0.2,
		"max_tokens":      maxTokens,
		"response_format": map[string]string{"type": "json_object"},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", nil, err
	}

	endpoint := defaultEndpoint
	if v, ok := ctx.Value(endpointKey{}).(string); ok && v != "" {
		endpoint = v
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 0} // deadline comes from ctx

	res, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("openai: request failed: %w", err)
	}
	defer res.Body.Close()
	raw, err = io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", nil, fmt.Errorf("openai: read body: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", raw, fmt.Errorf("openai: http %d", res.StatusCode)
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", raw, fmt.Errorf("openai: malformed provider json: %w", err)
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", raw, fmt.Errorf("openai: empty assistant content")
	}
	return parsed.Choices[0].Message.Content, raw, nil
}

// WithEndpoint overrides the API URL (tests).
func WithEndpoint(ctx context.Context, url string) context.Context {
	return context.WithValue(ctx, endpointKey{}, url)
}

// DefaultTimeout is the live-call bound.
const DefaultTimeout = 45 * time.Second

type endpointKey struct{}
