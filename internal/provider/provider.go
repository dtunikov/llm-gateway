package provider

//go:generate minimock -i github.com/dmitrii/llm-gateway/internal/provider.Provider -o provider_mock.go -n ProviderMock -p provider

import (
	"context"

	"github.com/dmitrii/llm-gateway/api"
)

// Provider is the interface that all LLM providers must implement.
type Provider interface {
	// ChatCompletion creates a completion for the given chat conversation.
	ChatCompletion(ctx context.Context, req *api.ChatCompletionRequest) (*api.ChatCompletionResponse, error)
}
