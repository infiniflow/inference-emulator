# ollama-mock

An **Ollama native API emulator** written in Go + Gin. Every response is a fixed mock value, so no real model is ever involved. It is meant for local development, CI pipelines, and client integration tests.

## Features

- Emulates the core Ollama endpoints (`tags` / `version` / `ps` / `generate` / `chat` / `embed` / `embeddings` / `show` / `pull` / `push` / `create` / `copy` / `delete`)
- Streaming endpoints follow NDJSON (`application/x-ndjson`) strictly, writing and flushing line by line
- JSON field names are identical to the native Ollama API
- Optional API key authentication (disabled unless a key is configured)
- Errors always use the `{"error": "..."}` shape (400 / 401 / 404 / 405 included)
- Only Gin plus the standard library

## Run

```bash
go build ./...
go run .                          # listens on :11434 by default
OLLAMA_MOCK_PORT=11435 go run .   # custom port
```

Environment variables:

| Variable | Description | Default |
| --- | --- | --- |
| `OLLAMA_MOCK_PORT` | Listening port | `11434` |
| `OLLAMA_MOCK_STREAM_DELAY_MS` | Delay between two streamed chunks (milliseconds, `0` disables the wait) | `10` |
| `OLLAMA_MOCK_API_KEY` | Single API key; enabling it turns authentication on | (unset) |
| `OLLAMA_MOCK_API_KEYS` | Comma-separated list of API keys (merged with the above) | (unset) |
| `OLLAMA_MOCK_AUTH_EXEMPT` | Comma-separated route paths that stay open even when auth is on, e.g. `/api/version` | (none) |
| `OLLAMA_MOCK_ACCESS_LOG` | Per-request access log; set to `off` (or `0` / `false` / `no`) to silence it | on |

## API key authentication

Authentication is **off by default** (the native Ollama server has none). As soon as
`OLLAMA_MOCK_API_KEY` or `OLLAMA_MOCK_API_KEYS` is set, every `/api/*` route requires a key.

```bash
OLLAMA_MOCK_API_KEY=sk-mock-123 go run .
```

Accepted header forms (checked in this order):

1. `X-API-Key: sk-mock-123`
2. `Authorization: Bearer sk-mock-123`
3. `Authorization: sk-mock-123` (bare key)

```bash
curl -s http://localhost:11434/api/tags -H 'Authorization: Bearer sk-mock-123' | jq
curl -s http://localhost:11434/api/tags -H 'X-API-Key: sk-mock-123' | jq
```

A missing or wrong key returns `401` with the unified error payload:

```json
{"error":"unauthorized"}
```

To keep a health-check endpoint open:

```bash
OLLAMA_MOCK_API_KEY=sk-mock-123 OLLAMA_MOCK_AUTH_EXEMPT=/api/version go run .
```

## Access log

Every request is printed to stdout as a single structured line, enabled by default:

```
2026/09/18 12:21:39.134312 access method=POST path=/api/generate status=200 latency=51.475ms bytes=740 ip=::1 key=sk-m...3456 model=llama3.2:latest
```

Fields: `method`, `path`, `status`, `latency`, `bytes` (response size), `ip` (client),
`key` (masked API key, `-` when none), `model` (from the JSON request body, empty for GETs).

```bash
# Silence the access log
OLLAMA_MOCK_ACCESS_LOG=off go run .

# Write it to a file
go run . > access.log 2>&1

# Follow / filter it
tail -f access.log | grep 'path=/api/chat'
```

## Fixed mock data

- Models: `llama3.2:latest`, `nomic-embed-text:latest`, `gemma2:2b` (fixed digest / size / modified_at)
- `/api/generate` fixed reply: `That is a great question!`
- `/api/chat` fixed reply: `Hello! How can I help?`
- Embeddings: every input returns the same 768-dimension vector whose every dimension is `0.1`
- Version: `0.5.7`

## Endpoints

| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/tags` | Three fixed models |
| GET | `/api/version` | `{"version":"0.5.7"}` |
| GET | `/api/ps` | Three fixed "running" models (with `expires_at` / `size_vram`) |
| POST | `/api/generate` | NDJSON stream when `stream` is true or omitted |
| POST | `/api/chat` | Same, but the payload uses a `message` object |
| POST | `/api/embed` | `input` accepts a string or a []string, returns 768-dimension vectors |
| POST | `/api/embeddings` | Legacy endpoint, returns a single `embedding` |
| POST | `/api/show` | Fixed details per model, 404 for unknown models |
| POST | `/api/pull` | Progress stream; with `stream:false` only `success` is returned |
| POST | `/api/push` | `retrieving manifest` -> `pushing manifest` -> `success` |
| POST | `/api/create` | `reading model metadata` -> ... -> `success` |
| POST | `/api/copy` | 200 with an empty body |
| DELETE | `/api/delete` | 200 with an empty body |

## Verify with curl

```bash
curl http://localhost:11434/api/tags
curl http://localhost:11434/api/version

curl http://localhost:11434/api/generate -d '{"model":"llama3.2","prompt":"hi","stream":false}'
curl -N http://localhost:11434/api/generate -d '{"model":"llama3.2","prompt":"hi"}'

curl http://localhost:11434/api/chat \
  -d '{"model":"llama3.2","messages":[{"role":"user","content":"hi"}],"stream":false}'
curl -N http://localhost:11434/api/chat \
  -d '{"model":"llama3.2","messages":[{"role":"user","content":"hi"}]}'

curl http://localhost:11434/api/embed -d '{"model":"nomic-embed-text","input":["a","b"]}'
curl http://localhost:11434/api/embeddings -d '{"model":"nomic-embed-text","prompt":"a"}'
curl http://localhost:11434/api/show -d '{"model":"llama3.2:latest"}'
curl -N http://localhost:11434/api/pull -d '{"model":"llama3.2:latest"}'
```

## Tests

```bash
go build ./...
go test ./...
```

## Project layout

```
main.go
internal/config      # port and stream delay
internal/models      # request/response structs + fixed mock data
internal/handlers    # endpoint handlers
internal/router      # route registration
internal/stream      # NDJSON streaming helpers
internal/auth        # optional API key middleware
internal/logging     # per-request access log
internal/httputil    # unified error responses
tests/api_test.go    # API tests
```
