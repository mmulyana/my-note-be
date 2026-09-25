package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"

var ErrUpstreamFailed = errors.New("ai request failed")

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouter struct {
	apiKey string
	model  string
	client *http.Client
}

func newOpenRouter(apiKey, model string) *openRouter {
	return &openRouter{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 30 * time.Second,
			},
		},
	}
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Usage *struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (o *openRouter) stream(ctx context.Context, messages []message, onDelta func(string) error) (tokens int, err error) {
	body, err := json.Marshal(struct {
		Model    string    `json:"model"`
		Messages []message `json:"messages"`
		Stream   bool      `json:"stream"`
	}{o.model, messages, true})
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterURL, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := o.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		log.Printf("openrouter: request failed: %v", err)
		return 0, ErrUpstreamFailed
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		log.Printf("openrouter: status %d: %s", resp.StatusCode, strings.TrimSpace(string(detail)))
		return 0, ErrUpstreamFailed
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if data, ok := strings.CutPrefix(strings.TrimSpace(line), "data:"); ok {
			data = strings.TrimSpace(data)
			if data == "[DONE]" {
				return tokens, nil
			}

			var chunk streamChunk
			if json.Unmarshal([]byte(data), &chunk) == nil {
				if chunk.Error != nil {
					log.Printf("openrouter: stream error: %s", chunk.Error.Message)
					return tokens, ErrUpstreamFailed
				}
				if chunk.Usage != nil {
					tokens = chunk.Usage.TotalTokens
				}
				if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
					if cbErr := onDelta(chunk.Choices[0].Delta.Content); cbErr != nil {
						return tokens, cbErr
					}
				}
			}
		}

		if err != nil {
			if err == io.EOF {
				return tokens, nil
			}
			if ctx.Err() != nil {
				return tokens, ctx.Err()
			}
			log.Printf("openrouter: read failed: %v", err)
			return tokens, ErrUpstreamFailed
		}
	}
}
