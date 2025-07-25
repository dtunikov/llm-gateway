package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "config-*.yml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(`
server:
  port: "9000"
logging:
  level: "debug"
providers:
  - id: openai-test
    provider: openai
    config:
      api_key: "test-key"
      api_url: "http://test.url"
models:
  - id: test-model
    name: test-model-name
    provider: openai-test
    fallback: ["fallback-model"]
`)
	assert.NoError(t, err)
	tmpFile.Close()

	// Set CONFIG_PATH environment variable to the temporary file
	os.Setenv("CONFIG_PATH", tmpFile.Name())
	defer os.Unsetenv("CONFIG_PATH")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "9000", cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Len(t, cfg.Providers, 1)
	assert.Equal(t, "openai-test", cfg.Providers[0].ID)
	assert.Equal(t, ProviderOpenAI, cfg.Providers[0].Provider)
	openAIConfig, ok := cfg.Providers[0].Config.(*OpenAIProviderConfig)
	assert.True(t, ok)
	assert.Equal(t, "test-key", openAIConfig.APIKey)
	assert.Equal(t, "http://test.url", openAIConfig.APIUrl)
	assert.Len(t, cfg.Models, 1)
	assert.Equal(t, "test-model", cfg.Models[0].ID)
	assert.Equal(t, "test-model-name", cfg.Models[0].Name)
	assert.Equal(t, "openai-test", cfg.Models[0].Provider)
	assert.Equal(t, []string{"fallback-model"}, cfg.Models[0].Fallback)
}

func TestLoadConfigEnvOverride(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "config-*.yml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(`
server:
  port: "8080"
logging:
  level: "info"
`)
	assert.NoError(t, err)
	tmpFile.Close()

	// Set CONFIG_PATH and environment variables to override
	os.Setenv("CONFIG_PATH", tmpFile.Name())
	os.Setenv("SERVER_PORT", "8000")
	os.Setenv("LOG_LEVEL", "warn")
	defer os.Unsetenv("CONFIG_PATH")
	defer os.Unsetenv("SERVER_PORT")
	defer os.Unsetenv("LOG_LEVEL")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "8000", cfg.Server.Port)
	assert.Equal(t, "warn", cfg.Logging.Level)
}

func TestLoadConfigNotFound(t *testing.T) {
	// Ensure CONFIG_PATH is not set to a valid file
	os.Setenv("CONFIG_PATH", "nonexistent.yml")
	defer os.Unsetenv("CONFIG_PATH")

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}
