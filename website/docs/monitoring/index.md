---
sidebar_position: 5
---

# Monitoring

The LLM Gateway provides robust monitoring capabilities out-of-the-box, allowing you to observe and track your LLM usage in real-time.

## Prometheus Metrics

The gateway exposes a Prometheus-compatible endpoint at `/metrics` on the application's port (default: `http://localhost:8080/metrics`). This endpoint provides detailed metrics on token usage, broken down by model and provider.

Key metrics include:

*   `llm_gateway_prompt_tokens_total{model="<model_name>", provider="<provider_name>"}`: Total number of prompt tokens processed.
*   `llm_gateway_completion_tokens_total{model="<model_name>", provider="<provider_name>"}`: Total number of completion tokens generated.
*   `llm_gateway_total_tokens_total{model="<model_name>", provider="<provider_name>"}`: Total number of tokens (prompt + completion).

## Pre-configured Grafana Dashboard

For immediate visualization, the LLM Gateway comes with a pre-configured Grafana dashboard. When you run the application using the provided Docker Compose setup, Grafana is automatically set up with a dashboard that visualizes the key token usage metrics.

This dashboard provides charts for:

*   **Token Usage Rate:** See the rate of prompt, completion, and total tokens per model and provider.
*   **Cumulative Token Usage:** Track the total number of tokens used over time, categorized by model and provider.
*   **Total Tokens per Model:** A bar chart showing the total token count for each model.

To access the dashboard, simply navigate to Grafana (default: `http://localhost:3000`) and find the "LLM Gateway Token Usage" dashboard.

## Building Your Own Dashboards

The Prometheus metrics exposed by the LLM Gateway are not limited to the pre-configured dashboard. You can easily create your own custom dashboards in Grafana or any other Prometheus-compatible visualization tool.

By using the provided metrics as a data source, you can build visualizations tailored to your specific needs, such as:

*   Cost analysis dashboards (by combining token counts with pricing information).
*   Performance monitoring dashboards (by tracking latency or error rates, if you add such metrics).
*   Usage dashboards for specific users or applications.

This flexibility allows you to gain deep insights into your LLM usage and monitor the performance and cost of your applications.
