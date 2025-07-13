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
  test_provider:
    api_key: "test-key"
    api_url: "http://test.url"
models:
  test-model: test_provider
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
	assert.Equal(t, "test-key", cfg.Providers["test_provider"].APIKey)
	assert.Equal(t, "http://test.url", cfg.Providers["test_provider"].APIUrl)
	assert.Equal(t, "test_provider", cfg.Models["test-model"])
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
