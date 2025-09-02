package dummy

import (
	"context"
	"fmt"
	"time"

	"github.com/dmitrii/llm-gateway/api"
)

// DummyProvider is a dummy implementation of the Provider interface.
type DummyProvider struct{}

func NewDummyProvider() *DummyProvider {
	return &DummyProvider{}
}

// ChatCompletion creates a dummy completion for the given chat conversation.
func (dp *DummyProvider) ChatCompletion(ctx context.Context, req *api.ChatCompletionRequest) (*api.ChatCompletionResponse, error) {
	// Simulate some work and token usage
	time.Sleep(100 * time.Millisecond)

	promptTokens := len(req.Messages) * 5 // Arbitrary token count for dummy
	completionTokens := 10
	totalTokens := promptTokens + completionTokens

	content := &api.ChatMessage_Content{}
	content.FromChatMessageContent0("Hello! This is a dummy response.")
	resp := &api.ChatCompletionResponse{
		Id:      fmt.Sprintf("dummy-cmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: int(time.Now().Unix()),
		Model:   req.Model,
		Choices: []api.ChatCompletionChoice{
			{
				Index:        0,
				Message:      api.ChatMessage{Role: "assistant", Content: content},
				FinishReason: "stop",
			},
		},
		Usage: &api.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
		},
	}

	return resp, nil
}
