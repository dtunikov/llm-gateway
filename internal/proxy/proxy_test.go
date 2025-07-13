package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitrii/llm-gateway/internal/config"
	"github.com/dmitrii/llm-gateway/internal/provider"
	"github.com/dmitrii/llm-gateway/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProvider is a mock implementation of the Provider interface.
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) ChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*types.ChatCompletionResponse), args.Error(1)
}

func TestProxy_ChatCompletionsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock providers
	mockOpenAIProvider := new(MockProvider)
	mockGeminiProvider := new(MockProvider)
	mockDummyProvider := new(MockProvider)

	// Configure mock behavior
	mockOpenAIProvider.On("ChatCompletion", mock.Anything, mock.AnythingOfType("*types.ChatCompletionRequest")).Return(&types.ChatCompletionResponse{
		ID:    "openai-test-id",
		Model: "gpt-4.1",
		Usage: types.Usage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2},
	}, nil)

	mockGeminiProvider.On("ChatCompletion", mock.Anything, mock.AnythingOfType("*types.ChatCompletionRequest")).Return(&types.ChatCompletionResponse{
		ID:    "gemini-test-id",
		Model: "gemini-2.5-pro",
		Usage: types.Usage{PromptTokens: 2, CompletionTokens: 2, TotalTokens: 4},
	}, nil)
	mockGeminiProvider.On("ChatCompletion", mock.Anything, mock.AnythingOfType("*types.ChatCompletionRequest")).Return(&types.ChatCompletionResponse{
		ID:    "gemini-test-id",
		Model: "gemini-2.5-pro",
		Usage: types.Usage{PromptTokens: 2, CompletionTokens: 2, TotalTokens: 4},
	}, nil)

	mockDummyProvider.On("ChatCompletion", mock.Anything, mock.AnythingOfType("*types.ChatCompletionRequest")).Return(&types.ChatCompletionResponse{
		ID:    "dummy-test-id",
		Model: "dummy-model",
		Usage: types.Usage{PromptTokens: 3, CompletionTokens: 3, TotalTokens: 6},
	}, nil)

	// Create a config with model mappings
	cfg := &config.Config{
		Models: map[string]string{
			"gpt-4.1":        "openai",
			"gemini-2.5-pro": "gemini",
			"dummy-model":    "dummy",
		},
	}

	// Create a map of providers
	providers := map[string]provider.Provider{
		"openai": mockOpenAIProvider,
		"gemini": mockGeminiProvider,
		"dummy":  mockDummyProvider,
	}

	// Create the Proxy instance
	llmProxy := &Proxy{
		cfg:       cfg,
		providers: providers,
	}

	// Create a Gin router and register the handler
	r := gin.New()
	r.POST("/v1/chat/completions", llmProxy.ChatCompletionsHandler)

	// Test cases
	tests := []struct {
		name           string
		model          string
		expectedStatus int
		expectedID     string
	}{
		{name: "OpenAI Model", model: "gpt-4.1", expectedStatus: http.StatusOK, expectedID: "openai-test-id"},
		{name: "Gemini Model", model: "gemini-2.5-pro", expectedStatus: http.StatusOK, expectedID: "gemini-test-id"},
		{name: "Dummy Model", model: "dummy-model", expectedStatus: http.StatusOK, expectedID: "dummy-test-id"},
		{name: "Unknown Model", model: "unknown-model", expectedStatus: http.StatusBadRequest, expectedID: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := map[string]interface{}{
				"model": tt.model,
				"messages": []map[string]string{
					{"role": "user", "content": "test"},
				},
			}
			reqBodyBytes, _ := json.Marshal(reqBody)

			rec := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBodyBytes))
			req.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp types.ChatCompletionResponse
				json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.Equal(t, tt.expectedID, resp.ID)
			}
		})
	}

	// Verify that mock methods were called as expected
	mockOpenAIProvider.AssertExpectations(t)
	mockGeminiProvider.AssertExpectations(t)
	mockDummyProvider.AssertExpectations(t)
}
