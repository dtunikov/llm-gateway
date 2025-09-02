package proxy

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dmitrii/llm-gateway/api"
	"github.com/dmitrii/llm-gateway/internal/config"
	"github.com/dmitrii/llm-gateway/internal/errors"
	"github.com/dmitrii/llm-gateway/internal/provider"
	"github.com/dmitrii/llm-gateway/internal/provider/dummy"
	langchaincompatible "github.com/dmitrii/llm-gateway/internal/provider/langchain_compatible"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/huggingface"
	"github.com/tmc/langchaingo/llms/ollama"
	llmsopenai "github.com/tmc/langchaingo/llms/openai"
)

var (
	promptTokensTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_gateway_prompt_tokens_total",
			Help: "Total number of prompt tokens used",
		},
		[]string{"model", "provider"},
	)
	completionTokensTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_gateway_completion_tokens_total",
			Help: "Total number of completion tokens used",
		},
		[]string{"model", "provider"},
	)
	totalTokensTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_gateway_total_tokens_total",
			Help: "Total number of tokens used (prompt + completion)",
		},
		[]string{"model", "provider"},
	)
)

func init() {
	prometheus.MustRegister(promptTokensTotal)
	prometheus.MustRegister(completionTokensTotal)
	prometheus.MustRegister(totalTokensTotal)
}

// Proxy holds the configuration and initialized LLM providers.
type Proxy struct {
	cfg       *config.Config
	providers map[string]provider.Provider
}

// NewProxy creates a new Proxy instance and initializes all configured providers.
func NewProxy(cfg *config.Config) (*Proxy, error) {
	providers := make(map[string]provider.Provider)
	var err error

	for _, pCfg := range cfg.Providers {
		id := pCfg.ID
		if pCfg.Provider == config.ProviderDummy {
			providers[id] = dummy.NewDummyProvider()
			continue
		}

		var llm llms.Model
		switch pCfg.Provider {
		case config.ProviderAnthropic:
			anthropicCfg := pCfg.Config.(*config.AnthropicProviderConfig)
			llm, err = anthropic.New(
				anthropic.WithBaseURL(anthropicCfg.APIUrl),
				anthropic.WithToken(anthropicCfg.APIKey),
			)
		case config.ProviderAzureOpenAI:
			azureCfg := pCfg.Config.(*config.AzureOpenAIProviderConfig)
			llm, err = llmsopenai.New(
				llmsopenai.WithToken(azureCfg.APIKey),
				llmsopenai.WithBaseURL(azureCfg.APIUrl),
				llmsopenai.WithAPIVersion(azureCfg.ApiVersion),
				llmsopenai.WithAPIType(azureCfg.ApiType),
			)
		case config.ProviderOpenAI:
			openaiCfg := pCfg.Config.(*config.OpenAIProviderConfig)
			llm, err = llmsopenai.New(
				llmsopenai.WithToken(openaiCfg.APIKey),
				llmsopenai.WithBaseURL(openaiCfg.APIUrl),
				llmsopenai.WithAPIVersion(openaiCfg.ApiVersion),
				llmsopenai.WithOrganization(openaiCfg.OrgID),
			)

		case config.ProviderGemini:
			geminiCfg := pCfg.Config.(*config.GeminiProviderConfig)
			llm, err = googleai.New(
				context.Background(),
				googleai.WithAPIKey(geminiCfg.APIKey),
			)
		case config.ProviderVertexAI:
			vertexCfg := pCfg.Config.(*config.VertexAIProviderConfig)
			llm, err = googleai.New(
				context.Background(),
				googleai.WithCloudProject(vertexCfg.ProjectID),
				googleai.WithCloudLocation(vertexCfg.Location),
				googleai.WithCredentialsFile(vertexCfg.PathToCredsFile),
			)
		case config.ProviderHuggingFace:
			hfCfg := pCfg.Config.(*config.HuggingFaceProviderConfig)
			llm, err = huggingface.New(
				huggingface.WithToken(hfCfg.APIKey),
				huggingface.WithURL(hfCfg.APIUrl),
			)
		case config.ProviderOllama:
			ollamaCfg := pCfg.Config.(*config.OllamaProviderConfig)
			llm, err = ollama.New(
				ollama.WithServerURL(ollamaCfg.APIUrl),
			)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create LLM model for provider %s: %w", id, err)
		}
		providers[id] = langchaincompatible.NewLangchainProvider(llm)
	}

	return &Proxy{
		cfg:       cfg,
		providers: providers,
	}, nil
}

// ChatCompletionsHandler handles requests to the /v1/chat/completions endpoint.
func (p *Proxy) ChatCompletionsHandler(ctx context.Context, req api.ChatCompletionRequest) (*api.ChatCompletionResponse, error) {
	var modelConfig *config.ModelConfig
	for _, m := range p.cfg.Models {
		if m.ID == req.Model {
			modelConfig = m
			break
		}
	}

	if modelConfig == nil {
		return nil, errors.ErrNotFound.WithMessage("model not found in config")
	}

	modelsToTry := []string{req.Model}
	modelsToTry = append(modelsToTry, modelConfig.Fallback...)

	var resp *api.ChatCompletionResponse
	var err error

	for _, modelID := range modelsToTry {
		var currentModelConfig *config.ModelConfig
		for _, m := range p.cfg.Models {
			if m.ID == modelID {
				currentModelConfig = m
				break
			}
		}

		if currentModelConfig == nil {
			slog.Error("Fallback model not found in config", "model", modelID)
			continue // Try next model
		}

		providerName := currentModelConfig.Provider
		llmProvider, ok := p.providers[providerName]
		if !ok {
			slog.Error("Provider not found for model", "model", modelID, "provider", providerName)
			continue // Try next model
		}

		slog.Info("Sending request to provider", "model", currentModelConfig.Name, "provider", providerName)
		// Create a new request object for each attempt to avoid modifying the original
		attemptReq := req
		attemptReq.Model = currentModelConfig.Name

		resp, err = llmProvider.ChatCompletion(ctx, &attemptReq)
		if err != nil {
			slog.Error("Provider chat completion failed", "error", err, "model", currentModelConfig.Name, "provider", providerName)
			continue // Try next model
		}

		// Increment token usage metrics
		if resp.Usage.PromptTokens > 0 {
			promptTokensTotal.WithLabelValues(resp.Model, providerName).Add(float64(resp.Usage.PromptTokens))
		}
		if resp.Usage.CompletionTokens > 0 {
			completionTokensTotal.WithLabelValues(resp.Model, providerName).Add(float64(resp.Usage.CompletionTokens))
		}
		if resp.Usage.TotalTokens > 0 {
			totalTokensTotal.WithLabelValues(resp.Model, providerName).Add(float64(resp.Usage.TotalTokens))
		}

		return resp, nil
	}

	return nil, errors.ErrInternal.WithMessage("failed to get completion from any provider")
}
