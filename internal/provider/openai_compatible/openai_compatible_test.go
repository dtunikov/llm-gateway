package openai_compatible

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dmitrii/llm-gateway/internal/config"
	"github.com/dmitrii/llm-gateway/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestOpenAICompatibleProvider_ChatCompletion(t *testing.T) {
	// Mock server for OpenAI-compatible API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		var reqBody types.ChatCompletionRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)
		assert.Equal(t, "test-model", reqBody.Model)
		assert.Len(t, reqBody.Messages, 1)
		assert.Equal(t, "user", reqBody.Messages[0].Role)
		assert.Equal(t, "Hello", reqBody.Messages[0].Content)

		resp := types.ChatCompletionResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "test-model",
			Choices: []types.ChatCompletionChoice{
				{
					Index:        0,
					Message:      types.ChatMessage{Role: "assistant", Content: "Hi there!"},
					FinishReason: "stop",
				},
			},
			Usage: types.Usage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create provider config
	cfg := config.ProviderConfig{
		APIKey: "test-api-key",
		APIUrl: server.URL,
	}

	// Create provider instance
	op := NewOpenAICompatibleProvider("test-provider", cfg)

	// Create request
	req := &types.ChatCompletionRequest{
		Model: "test-model",
		Messages: []types.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	// Perform chat completion
	resp, err := op.ChatCompletion(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, "test-id", resp.ID)
	assert.Equal(t, "chat.completion", resp.Object)
	assert.Equal(t, "test-model", resp.Model)
	assert.Len(t, resp.Choices, 1)
	assert.Equal(t, "Hi there!", resp.Choices[0].Message.Content)
	assert.Equal(t, 10, resp.Usage.PromptTokens)
	assert.Equal(t, 20, resp.Usage.CompletionTokens)
	assert.Equal(t, 30, resp.Usage.TotalTokens)
}

func TestOpenAICompatibleProvider_ChatCompletionError(t *testing.T) {
	// Mock server returning an error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
	}))
	defer server.Close()

	cfg := config.ProviderConfig{
		APIKey: "test-api-key",
		APIUrl: server.URL,
	}
	op := NewOpenAICompatibleProvider("test-provider", cfg)
	req := &types.ChatCompletionRequest{Model: "test-model"}

	resp, err := op.ChatCompletion(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "test-provider chat completion failed: API returned non-200 status: 500, body: Internal Server Error")
}
