# LLM Gateway

## Overview

LLM Gateway is a lightweight, high-performance proxy service built in Go (Golang) that provides a unified OpenAI-compatible API endpoint for various Large Language Model (LLM) providers. It allows you to seamlessly switch between different LLMs (e.g., OpenAI, Google Gemini, Ollama) by simply changing the `model` parameter in your requests, without modifying your application code.

## Features

*   **Unified API:** Exposes a single OpenAI-compatible `/v1/chat/completions` endpoint.
*   **Multiple Provider Support:** Currently supports:
    *   OpenAI
    *   Google Gemini (via its OpenAI-compatible API)
    *   Ollama (via its OpenAI-compatible API)
    *   Dummy Provider (for testing and development)
*   **Flexible Configuration:** Configurable via a YAML file (`config.yml`) and environment variables.
*   **Structured Logging:** Utilizes `slog` for structured, machine-readable logs.
*   **Prometheus Metrics:** Exposes token usage metrics (`prompt_tokens_total`, `completion_tokens_total`, `total_tokens_total`) for each model and provider.
*   **Grafana Dashboard:** Includes a pre-configured Grafana dashboard for visualizing token usage.
*   **OpenAPI Specification:** Provides an OpenAPI 3.0 specification (`openapi.yaml`) accessible via a Swagger UI endpoint.
*   **Docker Support:** Easily deployable using Docker and Docker Compose.

## Getting Started

### Prerequisites

*   [Go](https://golang.org/doc/install) (1.24 or higher)
*   [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) (for containerized deployment)

### Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/dmitrii/llm-gateway.git
    cd llm-gateway
    ```
2.  **Download Go modules:**
    ```bash
    go mod tidy
    ```

### Running Locally

1.  **Configure the application:**
    Edit `config.yml` to set up your LLM providers and model mappings. For example:
    ```yaml
    server:
      port: 8080

    logging:
      level: info

    providers:
      openai:
        api_key: "your-openai-api-key"
        api_url: "https://api.openai.com/v1"
        is_openai_compatible: true
      gemini:
        api_key: "your-gemini-api-key"
        api_url: "https://generativelanguage.googleapis.com"
        is_openai_compatible: true
      ollama:
        api_url: "http://localhost:11434"
        is_openai_compatible: true

    models:
      gpt-4.1: openai
      gemini-2.5-pro: gemini
      llama2: ollama
      dummy-model: dummy
    ```

2.  **Run the application:**
    ```bash
    go run cmd/llm-gateway/main.go
    ```
    The gateway will start on `http://localhost:8080`.

### Running with Docker Compose (Recommended for Testing)

This setup includes the LLM Gateway, Prometheus (for metrics collection), and Grafana (for visualization).

1.  **Ensure your `config.yml` is set up** as described above.
2.  **Build and start the services:**
    ```bash
    docker compose up --build
    ```
    This will build the `llm-gateway` Docker image and start all services.

3.  **Access the services:**
    *   **LLM Gateway:** `http://localhost:8080`
    *   **Prometheus:** `http://localhost:9090`
    *   **Grafana:** `http://localhost:3000` (Login with `admin`/`admin`)

    The Grafana dashboard for token usage (`LLM Gateway Token Usage`) should be automatically provisioned.

## API Endpoints

The LLM Gateway exposes an OpenAI-compatible API endpoint.

### Chat Completions

*   **Endpoint:** `POST /v1/chat/completions`
*   **Request Body:** Adheres to the [OpenAI Chat Completion Request format](https://platform.openai.com/docs/api-reference/chat/create).
    ```json
    {
      "model": "your-configured-model-name",
      "messages": [
        {"role": "user", "content": "Hello, how are you?"}
      ],
      "temperature": 0.7,
      "max_tokens": 150
    }
    ```
*   **Response Body:** Adheres to the [OpenAI Chat Completion Response format](https://platform.openai.com/docs/api-reference/chat/object).

### OpenAPI Specification (Swagger UI)

Access the interactive API documentation:

*   `http://localhost:8080/swagger/`

## Metrics

The LLM Gateway exposes Prometheus metrics at `http://localhost:8080/metrics`.

Key metrics include:

*   `llm_gateway_prompt_tokens_total{model="<model_name>", provider="<provider_name>"}`: Total prompt tokens.
*   `llm_gateway_completion_tokens_total{model="<model_name>", provider="<provider_name>"}`: Total completion tokens.
*   `llm_gateway_total_tokens_total{model="<model_name>", provider="<provider_name>"}`: Total tokens (prompt + completion).

## Grafana Dashboard

A pre-configured Grafana dashboard (`grafana/dashboards/llm-gateway-tokens.json`) is provided to visualize token usage metrics. It includes charts for:

*   Prompt, Completion, and Total Tokens per Model/Provider (Rate)
*   Cumulative Total Tokens per Model/Provider
*   Cumulative Total Tokens per Model (Bar Chart)

## Configuration

The application is configured via `config.yml` and environment variables. Environment variables take precedence over YAML values.

| Key (`config.yml`) | Environment Variable | Description                                     | Default Value |
| :----------------- | :------------------- | :---------------------------------------------- | :------------ |
| `server.port`      | `SERVER_PORT`        | Port for the HTTP server.                       | `8080`        |
| `logging.level`    | `LOG_LEVEL`          | Logging level (`debug`, `info`, `warn`, `error`). | `info`        |
| `providers.<name>.api_key` | N/A | API key for the specific provider.              |               |
| `providers.<name>.api_url` | N/A` | Base API URL for the specific provider.         |               |
| `models.<model_name>` | N/A | Maps a custom model name to a provider name.    |               |

## Contributing

Contributions are welcome! Please feel free to open issues or submit pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/dtunikov/llm-gateway/blob/main/LICENSE) file for details.