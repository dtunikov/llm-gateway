package server

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"time"

	"github.com/dmitrii/llm-gateway/internal/config"
	"github.com/dmitrii/llm-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
}

func New(cfg *config.Config, logger *slog.Logger) (*gin.Engine, error) {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(loggingMiddleware(logger, []string{"/metrics"}))
	r.Use(metricsMiddleware())

	// Initialize proxy
	llmProxy, err := proxy.NewProxy(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy: %w", err)
	}

	// API handler
	v1 := r.Group("/v1")
	{
		v1.POST("/chat/completions", llmProxy.ChatCompletionsHandler)
	}

	// Read and process OpenAPI spec
	openAPITemplate, err := template.ParseFiles(cfg.OpenAPI.SpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec template: %w", err)
	}

	var openAPIBuf bytes.Buffer
	if err := openAPITemplate.Execute(&openAPIBuf, cfg.Server); err != nil {
		return nil, fmt.Errorf("failed to execute OpenAPI template: %w", err)
	}

	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(200, "application/x-yaml", openAPIBuf.Bytes())
	})
	r.Static("/swagger", cfg.OpenAPI.UiPath)
	slog.Info("Access Swagger UI", "url", cfg.Server.BaseURL+"/swagger/")

	// Metrics handler
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return r, nil
}

func loggingMiddleware(logger *slog.Logger, ignorePaths []string) gin.HandlerFunc {
	ignorePathsMap := make(map[string]struct{})
	for _, path := range ignorePaths {
		ignorePathsMap[path] = struct{}{}
	}
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		if _, ok := ignorePathsMap[c.Request.URL.Path]; !ok {
			logger.Info("request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", c.Writer.Status(),
				"time", fmt.Sprintf("%vms", time.Since(start).Milliseconds()),
				"ip", c.ClientIP(),
			)
		}
	}
}

func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path).Inc()
	}
}
