# AI Gateway

A fork of [envoyproxy/ai-gateway](https://github.com/envoyproxy/ai-gateway) — an Envoy-based gateway for routing and managing AI/LLM API traffic.

## Overview

AI Gateway provides a unified interface for routing requests to multiple AI providers (OpenAI, Anthropic, Google Gemini, Ollama, etc.) with features like:

- **Multi-provider routing**: Route requests to different LLM backends based on rules
- **Rate limiting**: Control request rates per user, model, or provider
- **Token tracking**: Monitor and limit token usage across providers
- **Failover**: Automatically retry failed requests against alternative backends
- **Authentication**: Manage API keys and credentials for multiple providers

## Prerequisites

- Go 1.22+
- Envoy proxy (see `.envoy-version` for the required version)
- Docker & Docker Compose (for local development)
- [Ollama](https://ollama.com/) (for local LLM inference)

## Getting Started

### Local Development with Ollama

1. Copy the Ollama environment file and adjust as needed:
   ```bash
   cp .env.ollama .env
   ```

2. Start the local stack:
   ```bash
   docker compose up
   ```

3. Send a test request:
   ```bash
   curl -X POST http://localhost:10000/v1/chat/completions \
     -H 'Content-Type: application/json' \
     -d '{
       "model": "llama3.2",
       "messages": [{"role": "user", "content": "Hello!"}]
     }'
   ```

> **Personal note:** I'm running `llama3.2` locally — updated the model name above to match.
> Run `ollama list` to see what you have pulled and adjust accordingly.
> I also have `mistral` pulled as a fallback: `ollama pull mistral`
> Tip: `ollama pull llama3.2:1b` is a much lighter download (~1GB) if you're on a constrained machine.

### Building from Source

```bash
git clone https://github.com/your-org/ai-gateway.git
cd ai-gateway
go build ./...
```

### Running Tests

```bash
go test ./...
```

> **Note:** Integration tests require a running Envoy instance and may be slow. To run only unit tests:
> ```bash
> go test $(go list ./... | grep -v /tests/integration)
> ```

## Configuration

AI Gateway is configured via Envoy's xDS API or static configuration files. See the `examples/` directory for sample configurations.

### Supported Providers

| Provider | Status |
|----------|--------|
| OpenAI | ✅ Supported |
| Anthropic | ✅ Supported |
| Google Gemini | ✅ Supported |
| AWS Bedrock | ✅ Supported |
| Ollama (local) | ✅ Supported |

## Architecture

```
Client → Envoy Proxy → AI Gateway Filter → LLM Provider
                            │
                            ├── Request transformation
                            ├── Auth injection
                            ├── Token counting
                            └── Response normalization
```

## Useful Commands

> **Personal cheatsheet** — commands I find myself running repeatedly:

```bash
# Tail gateway logs while sending requests
docker compose logs -f ai-gateway

# Check which Ollama models are available locally
ollama list

# Quick chat alias — usage: chat "Why is the sky blue?"
alias chat='f(){ curl -s -X POST http://localhost:10000/v1/chat/completions -H "Content-Type: application/json" -d "{\"model\":\"llama3.2\",\"messages\":[{\"role\":\"user\",\"content\":\"$1\"}]}" | jq -r ".choices[0].message.content"; }; f'

# Restart just the gateway container without bringing down Ollama
docker compose restart ai-gateway
```
