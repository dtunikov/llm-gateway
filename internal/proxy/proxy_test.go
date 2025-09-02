/*
Package proxy_test contains comprehensive unit tests for the proxy package.

Test Coverage:
- NewProxy function: Tests successful proxy creation with various configurations and error handling
- ChatCompletionsHandler: Tests the main request handling logic including:
  - Successful completion with proper model and provider mapping
  - Model not found scenarios
  - Provider not found scenarios
  - Fallback logic when primary provider fails
  - Error handling when all providers fail
  - Invalid fallback model configurations
  - Token metrics tracking with various usage scenarios

The tests use mock providers to isolate the proxy logic and validate the behavior
without requiring actual LLM provider connections.
*/

package proxy

import (
	"context"
	"errors"
	"testing"

	"github.com/dmitrii/llm-gateway/api"
	"github.com/dmitrii/llm-gateway/internal/config"
	internalerrors "github.com/dmitrii/llm-gateway/internal/errors"
	"github.com/dmitrii/llm-gateway/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create chat message content
func createChatContent(text string) *api.ChatMessage_Content {
	content := &api.ChatMessage_Content{}
	content.FromChatMessageContent0(text)
	return content
}

func TestNewProxy_Success(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
	}{
		{
			name: "empty config",
			config: &config.Config{
				Providers: []*config.ProviderConfig{},
				Models:    []*config.ModelConfig{},
			},
		},
		{
			name: "dummy provider only",
			config: &config.Config{
				Providers: []*config.ProviderConfig{
					{
						ID:       "dummy1",
						Provider: config.ProviderDummy,
						Config:   &config.DummyProviderConfig{},
					},
				},
				Models: []*config.ModelConfig{
					{
						ID:       "test-model",
						Name:     "test-model",
						Provider: "dummy1",
						Fallback: []string{},
					},
				},
			},
		},
		{
			name: "multiple dummy providers",
			config: &config.Config{
				Providers: []*config.ProviderConfig{
					{
						ID:       "dummy1",
						Provider: config.ProviderDummy,
						Config:   &config.DummyProviderConfig{},
					},
					{
						ID:       "dummy2",
						Provider: config.ProviderDummy,
						Config:   &config.DummyProviderConfig{},
					},
				},
				Models: []*config.ModelConfig{
					{
						ID:       "test-model-1",
						Name:     "test-model-1",
						Provider: "dummy1",
						Fallback: []string{"test-model-2"},
					},
					{
						ID:       "test-model-2",
						Name:     "test-model-2",
						Provider: "dummy2",
						Fallback: []string{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := NewProxy(tt.config)
			require.NoError(t, err)
			require.NotNil(t, proxy)
			assert.Equal(t, tt.config, proxy.cfg)
			assert.Equal(t, len(tt.config.Providers), len(proxy.providers))
		})
	}
}

func TestNewProxy_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		expectedErrMsg string
	}{
		{
			name: "invalid openai config",
			config: &config.Config{
				Providers: []*config.ProviderConfig{
					{
						ID:       "openai1",
						Provider: config.ProviderOpenAI,
						Config: &config.OpenAIProviderConfig{
							APIKey:     "", // invalid empty key
							APIUrl:     "https://api.openai.com",
							ApiVersion: "v1",
						},
					},
				},
				Models: []*config.ModelConfig{},
			},
			expectedErrMsg: "failed to create LLM model for provider openai1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := NewProxy(tt.config)
			assert.Nil(t, proxy)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErrMsg)
		})
	}
}

func TestChatCompletionsHandler_Success(t *testing.T) {
	// Create a mock provider
	mockProvider := provider.NewProviderMock(t)

	// Setup test configuration
	cfg := &config.Config{
		Models: []*config.ModelConfig{
			{
				ID:       "test-model",
				Name:     "actual-model-name",
				Provider: "test-provider",
				Fallback: []string{},
			},
		},
	}

	// Create proxy with mock provider
	proxy := &Proxy{
		cfg: cfg,
		providers: map[string]provider.Provider{
			"test-provider": mockProvider,
		},
	}

	// Create content for test request
	userContent := createChatContent("Hello, world!")

	// Test request
	req := api.ChatCompletionRequest{
		Model: "test-model",
		Messages: []api.ChatMessage{
			{
				Role:    api.ChatMessageRoleUser,
				Content: userContent,
			},
		},
	}

	// Create content for expected response
	assistantContent := createChatContent("Hello! How can I help you today?")

	// Expected response
	expectedResp := &api.ChatCompletionResponse{
		Id:      "test-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "actual-model-name",
		Choices: []api.ChatCompletionChoice{
			{
				Index: 0,
				Message: api.ChatMessage{
					Role:    api.ChatMessageRoleAssistant,
					Content: assistantContent,
				},
				FinishReason: api.ChatCompletionChoiceFinishReasonStop,
			},
		},
		Usage: &api.Usage{
			PromptTokens:     10,
			CompletionTokens: 15,
			TotalTokens:      25,
		},
	}

	// Create content for mock expectation
	mockUserContent := createChatContent("Hello, world!")

	// Setup mock expectation
	mockProvider.ChatCompletionMock.Expect(context.Background(), &api.ChatCompletionRequest{
		Model: "actual-model-name",
		Messages: []api.ChatMessage{
			{
				Role:    api.ChatMessageRoleUser,
				Content: mockUserContent,
			},
		},
	}).Return(expectedResp, nil)

	// Execute the handler
	ctx := context.Background()
	resp, err := proxy.ChatCompletionsHandler(ctx, req)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
}

func TestChatCompletionsHandler_ModelNotFound(t *testing.T) {
	cfg := &config.Config{
		Models: []*config.ModelConfig{},
	}

	proxy := &Proxy{
		cfg:       cfg,
		providers: map[string]provider.Provider{},
	}

	req := api.ChatCompletionRequest{
		Model: "non-existent-model",
		Messages: []api.ChatMessage{
			{
				Role:    api.ChatMessageRoleUser,
				Content: createChatContent("Hello"),
			},
		},
	}

	ctx := context.Background()
	resp, err := proxy.ChatCompletionsHandler(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, internalerrors.ErrNotFound.WithMessage("model not found in config"), err)
}

func TestChatCompletionsHandler_ProviderNotFound(t *testing.T) {
	cfg := &config.Config{
		Models: []*config.ModelConfig{
			{
				ID:       "test-model",
				Name:     "actual-model-name",
				Provider: "non-existent-provider",
				Fallback: []string{},
			},
		},
	}

	proxy := &Proxy{
		cfg:       cfg,
		providers: map[string]provider.Provider{},
	}

	req := api.ChatCompletionRequest{
		Model: "test-model",
		Messages: []api.ChatMessage{
			{
				Role:    api.ChatMessageRoleUser,
				Content: createChatContent("Hello"),
			},
		},
	}

	ctx := context.Background()
	resp, err := proxy.ChatCompletionsHandler(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, internalerrors.ErrInternal.WithMessage("failed to get completion from any provider"), err)
}

func TestChatCompletionsHandler_FallbackSuccess(t *testing.T) {
	// Create mock providers
	mockProvider1 := provider.NewProviderMock(t)
	mockProvider2 := provider.NewProviderMock(t)

	// Setup test configuration with fallback
	cfg := &config.Config{
		Models: []*config.ModelConfig{
			{
				ID:       "test-model",
				Name:     "primary-model",
				Provider: "provider1",
				Fallback: []string{"fallback-model"},
			},
			{
				ID:       "fallback-model",
				Name:     "backup-model",
				Provider: "provider2",
				Fallback: []string{},
			},
		},
	}

	// Create proxy with mock providers
	proxy := &Proxy{
		cfg: cfg,
		providers: map[string]provider.Provider{
			"provider1": mockProvider1,
			"provider2": mockProvider2,
		},
	}

	// Test request
	req := api.ChatCompletionRequest{
		Model: "test-model",
		Messages: []api.ChatMessage{
			{
				Role:    api.ChatMessageRoleUser,
				Content: createChatContent("Hello, world!"),
			},
		},
	}

	// Expected response from fallback provider
	expectedResp := &api.ChatCompletionResponse{
		Id:      "fallback-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "backup-model",
		Choices: []api.ChatCompletionChoice{
			{
				Index: 0,
				Message: api.ChatMessage{
					Role:    api.ChatMessageRoleAssistant,
					Content: createChatContent("Hello from fallback!"),
				},
				FinishReason: api.ChatCompletionChoiceFinishReasonStop,
			},
		},
		Usage: &api.Usage{
			PromptTokens:     8,
			CompletionTokens: 12,
			TotalTokens:      20,
		},
	}

	// Setup mock expectations
	// Primary provider fails
	mockProvider1.ChatCompletionMock.Expect(context.Background(), &api.ChatCompletionRequest{
		Model: "primary-model",
		Messages: []api.ChatMessage{
			{
				Role:    api.ChatMessageRoleUser,
				Content: createChatContent("Hello, world!"),
			},
		},
	}).Return(nil, errors.New("primary provider failed"))

	// Fallback provider succeeds
	mockProvider2.ChatCompletionMock.Expect(context.Background(), &api.ChatCompletionRequest{
		Model: "backup-model",
		Messages: []api.ChatMessage{
			{
				Role:    api.ChatMessageRoleUser,
				Content: createChatContent("Hello, world!"),
			},
		},
	}).Return(expectedResp, nil)

	// Execute the handler
	ctx := context.Background()
	resp, err := proxy.ChatCompletionsHandler(ctx, req)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
}

func TestChatCompletionsHandler_AllProvidersFail(t *testing.T) {
	// Create mock providers
	mockProvider1 := provider.NewProviderMock(t)
	mockProvider2 := provider.NewProviderMock(t)

	// Setup test configuration with fallback
	cfg := &config.Config{
		Models: []*config.ModelConfig{
			{
				ID:       "test-model",
				Name:     "primary-model",
				Provider: "provider1",
				Fallback: []string{"fallback-model"},
			},
			{
				ID:       "fallback-model",
				Name:     "backup-model",
				Provider: "provider2",
				Fallback: []string{},
			},
		},
	}

	// Create proxy with mock providers
	proxy := &Proxy{
		cfg: cfg,
		providers: map[string]provider.Provider{
			"provider1": mockProvider1,
			"provider2": mockProvider2,
		},
	}

	// Test request
	req := api.ChatCompletionRequest{
		Model: "test-model",
		Messages: []api.ChatMessage{
			{
				Role:    api.ChatMessageRoleUser,
				Content: createChatContent("Hello, world!"),
			},
		},
	}

	// Setup mock expectations - both providers fail
	mockProvider1.ChatCompletionMock.Expect(context.Background(), &api.ChatCompletionRequest{
		Model: "primary-model",
		Messages: []api.ChatMessage{
			{
				Role:    api.ChatMessageRoleUser,
				Content: createChatContent("Hello, world!"),
			},
		},
	}).Return(nil, errors.New("primary provider failed"))

	mockProvider2.ChatCompletionMock.Expect(context.Background(), &api.ChatCompletionRequest{
		Model: "backup-model",
		Messages: []api.ChatMessage{
			{
				Role:    api.ChatMessageRoleUser,
				Content: createChatContent("Hello, world!"),
			},
		},
	}).Return(nil, errors.New("fallback provider failed"))

	// Execute the handler
	ctx := context.Background()
	resp, err := proxy.ChatCompletionsHandler(ctx, req)

	// Verify results
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, internalerrors.ErrInternal.WithMessage("failed to get completion from any provider"), err)
}

func TestChatCompletionsHandler_FallbackModelNotFound(t *testing.T) {
	// Create mock provider
	mockProvider1 := provider.NewProviderMock(t)

	// Setup test configuration with invalid fallback model
	cfg := &config.Config{
		Models: []*config.ModelConfig{
			{
				ID:       "test-model",
				Name:     "primary-model",
				Provider: "provider1",
				Fallback: []string{"non-existent-fallback"},
			},
		},
	}

	// Create proxy with mock provider
	proxy := &Proxy{
		cfg: cfg,
		providers: map[string]provider.Provider{
			"provider1": mockProvider1,
		},
	}

	// Test request
	req := api.ChatCompletionRequest{
		Model: "test-model",
		Messages: []api.ChatMessage{
			{
				Role:    api.ChatMessageRoleUser,
				Content: createChatContent("Hello, world!"),
			},
		},
	}

	// Setup mock expectation - primary provider fails
	mockProvider1.ChatCompletionMock.Expect(context.Background(), &api.ChatCompletionRequest{
		Model: "primary-model",
		Messages: []api.ChatMessage{
			{
				Role:    api.ChatMessageRoleUser,
				Content: createChatContent("Hello, world!"),
			},
		},
	}).Return(nil, errors.New("primary provider failed"))

	// Execute the handler
	ctx := context.Background()
	resp, err := proxy.ChatCompletionsHandler(ctx, req)

	// Verify results
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, internalerrors.ErrInternal.WithMessage("failed to get completion from any provider"), err)
}

func TestChatCompletionsHandler_TokenMetrics(t *testing.T) {
	// Create a mock provider
	mockProvider := provider.NewProviderMock(t)

	// Setup test configuration
	cfg := &config.Config{
		Models: []*config.ModelConfig{
			{
				ID:       "test-model",
				Name:     "actual-model-name",
				Provider: "test-provider",
				Fallback: []string{},
			},
		},
	}

	// Create proxy with mock provider
	proxy := &Proxy{
		cfg: cfg,
		providers: map[string]provider.Provider{
			"test-provider": mockProvider,
		},
	}

	// Test request
	req := api.ChatCompletionRequest{
		Model: "test-model",
		Messages: []api.ChatMessage{
			{
				Role:    api.ChatMessageRoleUser,
				Content: createChatContent("Hello, world!"),
			},
		},
	}

	// Test cases for different token counts
	testCases := []struct {
		name  string
		usage *api.Usage
	}{
		{
			name: "all tokens present",
			usage: &api.Usage{
				PromptTokens:     10,
				CompletionTokens: 15,
				TotalTokens:      25,
			},
		},
		{
			name: "zero tokens",
			usage: &api.Usage{
				PromptTokens:     0,
				CompletionTokens: 0,
				TotalTokens:      0,
			},
		},
		{
			name: "partial tokens",
			usage: &api.Usage{
				PromptTokens:     5,
				CompletionTokens: 0,
				TotalTokens:      10,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Expected response with specific token usage
			expectedResp := &api.ChatCompletionResponse{
				Id:      "test-id",
				Object:  "chat.completion",
				Created: 1234567890,
				Model:   "actual-model-name",
				Choices: []api.ChatCompletionChoice{
					{
						Index: 0,
						Message: api.ChatMessage{
							Role:    api.ChatMessageRoleAssistant,
							Content: createChatContent("Hello! How can I help you today?"),
						},
						FinishReason: api.ChatCompletionChoiceFinishReasonStop,
					},
				},
				Usage: tc.usage,
			}

			// Setup mock expectation
			mockProvider.ChatCompletionMock.Expect(context.Background(), &api.ChatCompletionRequest{
				Model: "actual-model-name",
				Messages: []api.ChatMessage{
					{
						Role:    api.ChatMessageRoleUser,
						Content: createChatContent("Hello, world!"),
					},
				},
			}).Return(expectedResp, nil)

			// Execute the handler
			ctx := context.Background()
			resp, err := proxy.ChatCompletionsHandler(ctx, req)

			// Verify results
			require.NoError(t, err)
			assert.Equal(t, expectedResp, resp)
			assert.Equal(t, tc.usage, resp.Usage)
		})
	}
}
