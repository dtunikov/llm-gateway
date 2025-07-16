---
sidebar_position: 2
---

# Getting Started

This guide will walk you through the process of setting up and running the LLM Gateway.

## Prerequisites

*   [Go](https://golang.org/doc/install) (1.24 or higher)
*   [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) (for containerized deployment)

## Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/dtunikov/llm-gateway.git
    cd llm-gateway
    ```
2.  **Download Go modules:**
    ```bash
    go mod tidy
    ```

## Running Locally

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

## Running with Docker Compose (Recommended)

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
