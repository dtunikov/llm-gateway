package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/tmc/langchaingo/llms/openai"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
// It is loaded from a YAML file and/or environment variables.
// `yaml` and `env` tags are used to specify the mapping.
// `env-default` provides default values.
// `env-required` marks a field as mandatory.
// `env-description` provides a description for the environment variable.

type ProviderName string

const (
	ProviderOpenAI      ProviderName = "openai"
	ProviderAzureOpenAI ProviderName = "azure_openai"
	ProviderAnthropic   ProviderName = "anthropic"
	ProviderGemini      ProviderName = "gemini"
	ProviderOllama      ProviderName = "ollama"
	ProviderHuggingFace ProviderName = "huggingface"
	ProviderVertexAI    ProviderName = "vertex_ai"
	ProviderDummy       ProviderName = "dummy"
)

var configTypeFactories = map[ProviderName]func() ProviderConfigInterface{
	ProviderOpenAI:      func() ProviderConfigInterface { return &OpenAIProviderConfig{} },
	ProviderAzureOpenAI: func() ProviderConfigInterface { return &AzureOpenAIProviderConfig{} },
	ProviderAnthropic:   func() ProviderConfigInterface { return &AnthropicProviderConfig{} },
	ProviderGemini:      func() ProviderConfigInterface { return &GeminiProviderConfig{} },
	ProviderOllama:      func() ProviderConfigInterface { return &OllamaProviderConfig{} },
	ProviderHuggingFace: func() ProviderConfigInterface { return &HuggingFaceProviderConfig{} },
	ProviderVertexAI:    func() ProviderConfigInterface { return &VertexAIProviderConfig{} },
	ProviderDummy:       func() ProviderConfigInterface { return &DummyProviderConfig{} },
}

type Config struct {
	Server    ServerConfig         `yaml:"server"`
	Logging   LoggingConfig        `yaml:"logging"`
	Providers []*ProviderConfig    `yaml:"providers"`
	Models    []*ModelConfig       `yaml:"models"`
	OpenAPI   OpenApiConfig        `yaml:"openapi"`
}

type OpenApiConfig struct {
	SpecPath string `yaml:"spec_path" env:"OPENAPI_SPEC_PATH" env-default:"./api/openapi.yaml"`
	UiPath   string `yaml:"ui_path" env:"OPENAPI_UI_PATH" env-default:"./api/swagger-ui"`
}

// ServerConfig represents the server configuration.

type ServerConfig struct {
	Port    string `yaml:"port" env:"SERVER_PORT" env-default:"8080"`
	BaseURL string `yaml:"base_url" env:"SERVER_BASE_URL" env-default:"http://localhost:8080"`
}

// LoggingConfig represents the logging configuration.

type LoggingConfig struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
}

// ModelConfig represents the configuration for a specific model.
type ModelConfig struct {
	ID       string   `yaml:"id"`
	Name     string   `yaml:"name"`
	Provider string   `yaml:"provider"`
	Fallback []string `yaml:"fallback"`
}

type OpenAIProviderConfig struct {
	APIKey     string `yaml:"api_key" env:"OPENAI_API_KEY"`
	APIUrl     string `yaml:"api_url" env:"OPENAI_API_URL" env-default:"https://api.openai.com"`
	OrgID      string `yaml:"org_id" env:"OPENAI_ORG_ID"`
	ApiVersion string `yaml:"api_version" env:"OPENAI_API_VERSION" env-default:"v1"`
}

type AzureOpenAIProviderConfig struct {
	APIKey     string         `yaml:"api_key" env:"AZURE_OPENAI_API_KEY"`
	APIUrl     string         `yaml:"api_url" env:"AZURE_OPENAI_API_URL" env-default:"https://{your-custom-endpoint}.openai.azure.com/"`
	ApiVersion string         `yaml:"api_version" env:"AZURE_OPENAI_API_VERSION" env-default:"v1"`
	ApiType    openai.APIType `yaml:"api_type" env:"AZURE_OPENAI_API_TYPE" env-default:"AZURE"`
}

type AnthropicProviderConfig struct {
	APIKey string `yaml:"api_key" env:"ANTHROPIC_API_KEY"`
	APIUrl string `yaml:"api_url" env:"ANTHROPIC_API_URL" env-default:"https://api.anthropic.com/v1"`
}

type GeminiProviderConfig struct {
	APIKey        string `yaml:"api_key" env:"GEMINI_API_KEY"`
	CloudLocation string `yaml:"cloud_location" env:"GEMINI_CLOUD_LOCATION" env-default:"us-central1"`
}

type OllamaProviderConfig struct {
	APIUrl string `yaml:"api_url" env:"OLLAMA_API_URL" env-default:"http://localhost:11434"`
}

type HuggingFaceProviderConfig struct {
	APIKey string `yaml:"api_key" env:"HF_TOKEN"`
	APIUrl string `yaml:"api_url" env:"HF_API_URL" env-default:"https://api-inference.huggingface.co"`
}

type DummyProviderConfig struct{}

type VertexAIProviderConfig struct {
	ProjectID       string `yaml:"project_id" env:"VERTEX_AI_PROJECT_ID"`
	Location        string `yaml:"location" env:"VERTEX_AI_LOCATION" env-default:"us-central1"`
	PathToCredsFile string `yaml:"path_to_creds_file" env:"VERTEX_AI_CREDS_FILE"`
}

func (OpenAIProviderConfig) isProviderConfig()      {}
func (AzureOpenAIProviderConfig) isProviderConfig() {}
func (AnthropicProviderConfig) isProviderConfig()   {}
func (GeminiProviderConfig) isProviderConfig()      {}
func (OllamaProviderConfig) isProviderConfig()      {}
func (HuggingFaceProviderConfig) isProviderConfig() {}
func (VertexAIProviderConfig) isProviderConfig()    {}
func (DummyProviderConfig) isProviderConfig()       {}

type ProviderConfigInterface interface {
	isProviderConfig()
}

type ProviderConfig struct {
	ID       string                  `yaml:"id"`
	Provider ProviderName            `yaml:"provider"`
	Config   ProviderConfigInterface `yaml:"-"`
	Raw      yaml.Node               `yaml:"config"`
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
	for _, providerCfg := range cfg.Providers {
		if providerCfg.ID == "" {
			return nil, fmt.Errorf("provider id is required in config for provider with parameters: %v", providerCfg.Config)
		}

		factory, ok := configTypeFactories[providerCfg.Provider]
		if !ok {
			return nil, fmt.Errorf("no config type factory found for provider %q", providerCfg.Provider)
		}

		providerCfg.Config = factory()
		if err := providerCfg.Raw.Decode(providerCfg.Config); err != nil {
			return nil, fmt.Errorf("failed to decode provider config for %q: %w", providerCfg.Provider, err)
		}
		fmt.Printf("Loaded provider %q with config: %+v\n", providerCfg.Provider, providerCfg.Config)
	}

	return &cfg, nil
}
