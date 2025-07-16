---
sidebar_position: 4
---

# API Reference

The LLM Gateway exposes an OpenAI-compatible API endpoint.

## Chat Completions

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

## OpenAPI Specification (Swagger UI)

Access the interactive API documentation:

*   `http://localhost:8080/swagger/`
