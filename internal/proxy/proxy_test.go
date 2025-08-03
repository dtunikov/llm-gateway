package proxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitrii/llm-gateway/internal/config"
	"github.com/dmitrii/llm-gateway/internal/provider"
	"github.com/dmitrii/llm-gateway/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
)

func TestProxy_ChatCompletionsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a config with model mappings
	cfg := &config.Config{
		Models: []*config.ModelConfig{
			{ID: "gpt-4.1", Name: "gpt-4.1", Provider: "openai", Fallback: []string{"gemini-2.5-pro"}},
			{ID: "gemini-2.5-pro", Name: "gemini-2.5-pro", Provider: "gemini"},
			{ID: "dummy-model", Name: "dummy-model", Provider: "dummy"},
		},
	}

	// Test cases
	tests := []struct {
		name           string
		model          string
		expectedStatus int
		expectedID     string
		mockSetup      func(mockOpenAI, mockGemini, mockDummy *provider.ProviderMock)
	}{
		{
			name:           "OpenAI Model",
			model:          "gpt-4.1",
			expectedStatus: http.StatusOK,
			expectedID:     "openai-test-id",
			mockSetup: func(mockOpenAI, mockGemini, mockDummy *provider.ProviderMock) {
				mockOpenAI.ChatCompletionMock.Return(&types.ChatCompletionResponse{
					ID:    "openai-test-id",
					Model: "gpt-4.1",
					Usage: types.Usage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2},
				}, nil)
			},
		},
		{
			name:           "Gemini Model",
			model:          "gemini-2.5-pro",
			expectedStatus: http.StatusOK,
			expectedID:     "gemini-test-id",
			mockSetup: func(mockOpenAI, mockGemini, mockDummy *provider.ProviderMock) {
				mockGemini.ChatCompletionMock.Return(&types.ChatCompletionResponse{
					ID:    "gemini-test-id",
					Model: "gemini-2.5-pro",
					Usage: types.Usage{PromptTokens: 2, CompletionTokens: 2, TotalTokens: 4},
				}, nil)
			},
		},
		{
			name:           "Dummy Model",
			model:          "dummy-model",
			expectedStatus: http.StatusOK,
			expectedID:     "dummy-test-id",
			mockSetup: func(mockOpenAI, mockGemini, mockDummy *provider.ProviderMock) {
				mockDummy.ChatCompletionMock.Return(&types.ChatCompletionResponse{
					ID:    "dummy-test-id",
					Model: "dummy-model",
					Usage: types.Usage{PromptTokens: 3, CompletionTokens: 3, TotalTokens: 6},
				}, nil)
			},
		},
		{
			name:           "Unknown Model",
			model:          "unknown-model",
			expectedStatus: http.StatusBadRequest,
			expectedID:     "",
			mockSetup:      func(mockOpenAI, mockGemini, mockDummy *provider.ProviderMock) {},
		},
		{
			name:           "Fallback Model",
			model:          "gpt-4.1",
			expectedStatus: http.StatusOK,
			expectedID:     "gemini-test-id",
			mockSetup: func(mockOpenAI, mockGemini, mockDummy *provider.ProviderMock) {
				mockOpenAI.ChatCompletionMock.Return(nil, errors.New("openai error"))
				mockGemini.ChatCompletionMock.Return(&types.ChatCompletionResponse{
					ID:    "gemini-test-id",
					Model: "gemini-2.5-pro",
					Usage: types.Usage{PromptTokens: 2, CompletionTokens: 2, TotalTokens: 4},
				}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mock controller for each test
			mc := minimock.NewController(t)

			// Create fresh mock providers for each test
			mockOpenAIProvider := provider.NewProviderMock(mc)
			mockGeminiProvider := provider.NewProviderMock(mc)
			mockDummyProvider := provider.NewProviderMock(mc)

			// Setup mocks for each test
			tt.mockSetup(mockOpenAIProvider, mockGeminiProvider, mockDummyProvider)

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
}
