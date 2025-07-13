package openai_compatible

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dmitrii/llm-gateway/internal/client"
	"github.com/dmitrii/llm-gateway/internal/config"
	"github.com/dmitrii/llm-gateway/internal/types"
)

// OpenAICompatibleProvider implements the provider.Provider interface for OpenAI-compatible APIs.
type OpenAICompatibleProvider struct {
	providerName string
	APIKey       string
	APIUrl       string
	Client       *http.Client
}

// NewOpenAICompatibleProvider creates a new OpenAICompatibleProvider.
func NewOpenAICompatibleProvider(providerName string, cfg config.ProviderConfig) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{
		providerName: providerName,
		APIKey:       cfg.APIKey,
		APIUrl:       cfg.APIUrl,
		Client:       &http.Client{},
	}
}

// Name returns the name of the provider.
func (op *OpenAICompatibleProvider) Name() string {
	return op.providerName
}

// ChatCompletion creates a completion for the given chat conversation using the OpenAI-compatible API.
func (op *OpenAICompatibleProvider) ChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error) {
	url := fmt.Sprintf("%s/v1/chat/completions", op.APIUrl)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	if op.APIKey != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", op.APIKey)
	}

	var apiResp types.ChatCompletionResponse
	if err := client.DoRequest(ctx, op.Client, "POST", url, headers, req, &apiResp); err != nil {
		return nil, fmt.Errorf("%s chat completion failed: %w", op.providerName, err)
	}

	return &apiResp, nil
}
