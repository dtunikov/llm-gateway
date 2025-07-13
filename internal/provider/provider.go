package provider

import (
	"context"

	"github.com/dmitrii/llm-gateway/internal/types"
)

// Provider is the interface that all LLM providers must implement.
type Provider interface {
	// Name returns the name of the provider.
	Name() string

	// ChatCompletion creates a completion for the given chat conversation.
	ChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error)
}
