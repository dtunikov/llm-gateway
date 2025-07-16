package proxy

import (
	"log/slog"
	"net/http"

	"github.com/dmitrii/llm-gateway/internal/config"
	"github.com/dmitrii/llm-gateway/internal/provider"
	"github.com/dmitrii/llm-gateway/internal/provider/dummy"
	"github.com/dmitrii/llm-gateway/internal/provider/openai_compatible"
	"github.com/dmitrii/llm-gateway/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
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

	providers["dummy"] = &dummy.DummyProvider{}
	for name, providerCfg := range cfg.Providers {
		if providerCfg.IsOpenAICompatible {
			providers[name] = openai_compatible.NewOpenAICompatibleProvider(name, providerCfg)
		}
	}

	return &Proxy{
		cfg:       cfg,
		providers: providers,
	}, nil
}

// ChatCompletionsHandler handles requests to the /v1/chat/completions endpoint.
func (p *Proxy) ChatCompletionsHandler(c *gin.Context) {
	var req types.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	modelConfig, ok := p.cfg.Models[req.Model]
	if !ok {
		slog.Error("Model not found in config", "model", req.Model)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Model not found"})
		return
	}

	modelsToTry := []string{req.Model}
	modelsToTry = append(modelsToTry, modelConfig.Fallback...)

	var resp *types.ChatCompletionResponse
	var err error

	for _, modelName := range modelsToTry {
		currentModelConfig, ok := p.cfg.Models[modelName]
		if !ok {
			slog.Error("Fallback model not found in config", "model", modelName)
			continue
		}
		providerName := currentModelConfig.Provider
		llmProvider, ok := p.providers[providerName]
		if !ok {
			slog.Error("Provider not found for model", "model", modelName, "provider", providerName)
			continue // Try next model
		}

		slog.Info("Sending request to provider", "model", modelName, "provider", providerName)
		// Create a new request object for each attempt to avoid modifying the original
		attemptReq := req
		attemptReq.Model = modelName

		resp, err = llmProvider.ChatCompletion(c.Request.Context(), &attemptReq)
		if err == nil {
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
			c.JSON(http.StatusOK, resp)
			return
		}
		slog.Error("Provider chat completion failed", "error", err, "model", modelName, "provider", providerName)
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get completion from any provider"})
}
