---
sidebar_position: 3
---

# Configuration

The application is configured via `config.yml` and environment variables. Environment variables take precedence over YAML values.

| Key (`config.yml`) | Environment Variable | Description                                     | Default Value |
| :----------------- | :------------------- | :---------------------------------------------- | :------------ |
| `server.port`      | `SERVER_PORT`        | Port for the HTTP server.                       | `8080`        |
| `logging.level`    | `LOG_LEVEL`          | Logging level (`debug`, `info`, `warn`, `error`). | `info`        |
| `providers.<name>.api_key` | N/A | API key for the specific provider.              |               |
| `providers.<name>.api_url` | N/A` | Base API URL for the specific provider.         |               |
| `models.<model_name>` | N/A | Maps a custom model name to a provider name.    |               |
