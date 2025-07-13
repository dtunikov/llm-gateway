package dummy

import (
	"context"
	"fmt"
	"time"

	"github.com/dmitrii/llm-gateway/internal/types"
)

// DummyProvider is a dummy implementation of the Provider interface.
type DummyProvider struct{}

// Name returns the name of the dummy provider.
func (dp *DummyProvider) Name() string {
	return "dummy"
}

// ChatCompletion creates a dummy completion for the given chat conversation.
func (dp *DummyProvider) ChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error) {
	// Simulate some work and token usage
	time.Sleep(100 * time.Millisecond)

	promptTokens := len(req.Messages) * 5 // Arbitrary token count for dummy
	completionTokens := 10
	totalTokens := promptTokens + completionTokens

	resp := &types.ChatCompletionResponse{
		ID:      fmt.Sprintf("dummy-cmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []types.ChatCompletionChoice{
			{
				Index:        0,
				Message:      types.ChatMessage{Role: "assistant", Content: "This is a dummy response."},
				FinishReason: "stop",
			},
		},
		Usage: types.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
		},
	}

	return resp, nil
}
