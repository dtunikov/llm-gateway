package main

import (
	"log/slog"

	"github.com/dmitrii/llm-gateway/internal/config"
	"github.com/dmitrii/llm-gateway/internal/log"
	"github.com/dmitrii/llm-gateway/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		return
	}

	logger := log.New(cfg.Logging.Level)
	slog.SetDefault(logger)

	slog.Info("Starting LLM Gateway", "port", cfg.Server.Port)

	r, err := server.New(cfg, logger)
	if err != nil {
		slog.Error("Failed to init server", "error", err)
		return
	}

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
