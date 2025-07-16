package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config represents the application configuration.
// It is loaded from a YAML file and/or environment variables.
// `yaml` and `env` tags are used to specify the mapping.
// `env-default` provides default values.
// `env-required` marks a field as mandatory.
// `env-description` provides a description for the environment variable.

type Config struct {
	Server    ServerConfig               `yaml:"server"`
	Logging   LoggingConfig              `yaml:"logging"`
	Providers map[string]ProviderConfig  `yaml:"providers"`
	Models    map[string]ModelConfig     `yaml:"models"`
	OpenAPI   OpenApiConfig              `yaml:"openapi"`
}

type OpenApiConfig struct {
	SpecPath string `yaml:"spec_path" env:"OPENAPI_SPEC_PATH" env-default:"./api/openapi.yaml"`
	UiPath   string `yaml:"ui_path" env:"OPENAPI_UI_PATH" env-default:"./api/swagger-ui"`
}

// ServerConfig represents the server configuration.

type ServerConfig struct {
	Port string `yaml:"port" env:"SERVER_PORT" env-default:"8080"`
}

// LoggingConfig represents the logging configuration.

type LoggingConfig struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
}

// ModelConfig represents the configuration for a specific model.
type ModelConfig struct {
	Provider string   `yaml:"provider"`
	Fallback []string `yaml:"fallback"`
}

// ProviderConfig represents the configuration for an LLM provider.

type ProviderConfig struct {
	APIKey             string `yaml:"api_key"`
	APIUrl             string `yaml:"api_url"`
	IsOpenAICompatible bool   `yaml:"is_openai_compatible" env-default:"true"`
}

// Load loads the configuration from a file and/or environment variables.
// The config file path is read from the `CONFIG_PATH` environment variable.
// If `CONFIG_PATH` is not set, it defaults to `config.yml`.

func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yml"
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
