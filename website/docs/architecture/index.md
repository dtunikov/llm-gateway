---
sidebar_position: 6
---

# Architecture

The LLM Gateway is designed with a modular and extensible architecture, centered around the following key components:

## Core Components

*   **Main Application (`cmd/llm-gateway/main.go`):** This is the entry point of the application. It's responsible for:
    *   Loading the configuration (`config.yml`).
    *   Initializing the logger.
    *   Creating and starting the HTTP server.

*   **HTTP Server (`internal/server/server.go`):** The server is built using the [Gin](https://gin-gonic.com/) web framework. It's responsible for:
    *   Handling incoming HTTP requests.
    *   Routing requests to the appropriate handlers.
    *   Serving the OpenAPI specification and Swagger UI.
    *   Exposing the Prometheus metrics endpoint.

*   **Proxy (`internal/proxy/proxy.go`):** The proxy is the core of the gateway. It's responsible for:
    *   Handling chat completion requests (`/v1/chat/completions`).
    *   Determining which provider to use based on the requested model.
    *   Forwarding the request to the appropriate provider.
    *   Handling model fallbacks if a provider fails.
    *   Recording Prometheus metrics for token usage.

*   **Providers (`internal/provider/`):** Providers are responsible for interacting with the different LLM APIs. The gateway uses a `Provider` interface to ensure that all providers have a consistent API. The following providers are currently implemented:
    *   **OpenAI-Compatible (`internal/provider/openai_compatible`):** A generic provider that can be used with any OpenAI-compatible API (e.g., OpenAI, Google Gemini, Ollama).
    *   **Dummy (`internal/provider/dummy`):** A simple provider for testing and development that returns a fixed response.

## Request Flow

1.  A user sends a chat completion request to the LLM Gateway's `/v1/chat/completions` endpoint.
2.  The Gin server receives the request and routes it to the `ChatCompletionsHandler` in the proxy.
3.  The proxy looks up the requested model in the configuration to determine which provider to use.
4.  The proxy forwards the request to the appropriate provider.
5.  The provider sends the request to the underlying LLM API.
6.  The provider receives the response from the LLM API and returns it to the proxy.
7.  The proxy records the token usage metrics and returns the response to the user.

This modular architecture makes it easy to add new providers and extend the functionality of the gateway.
