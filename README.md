# graceful-echo-server

[![CI](https://github.com/Anna1293/graceful-echo-server/actions/workflows/ci.yml/badge.svg)](https://github.com/Anna1293/graceful-echo-server/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev/)

Go echo backend with a **TLS reverse proxy** (Nginx replacement): HTTP/2, graceful shutdown, modular `internal/*` layout.

---

## Architecture

```text
Client --HTTPS :8443--> [ reverse proxy ] --HTTP :8080--> [ echo backend ]
                              |                                  |
                         TLS terminate                    GET /echo, /health
```

```mermaid
flowchart LR
  C[Client] -->|HTTPS 8443| P[proxy]
  P -->|HTTP 8080| E[echo]
```

---

## Features

- `GET /echo` — optional `msg` and `delay` query params
- Request cancellation via `r.Context()`
- Graceful shutdown on `SIGINT` / `SIGTERM`
- Reverse proxy on `net/http` with `X-Forwarded-*` headers
- TLS termination and HTTP/2 on `:8443`
- `GET /health` on backend and proxy (Docker healthcheck)
- Self-signed TLS certs on first run (or generate via script)

---

## Requirements

- Go 1.25+
- Optional: Docker, OpenSSL (for cert scripts)

---

## Quick start

```bash
go run .
```

Servers:

| Service | Address | Protocol |
|---------|---------|----------|
| Backend | `:8080` | HTTP |
| Proxy   | `:8443` | HTTPS (HTTP/2) |

### Examples

```bash
curl http://localhost:8080/health
curl "http://localhost:8080/echo?msg=hello&delay=0"
curl -k --http2 "https://localhost:8443/echo?msg=hello&delay=0"
```

`-k` skips verification of the self-signed certificate.

---

## Configuration (env)

| Variable | Default | Description |
|----------|---------|-------------|
| `BACKEND_ADDR` | `:8080` | Echo server listen address |
| `PROXY_ADDR` | `:8443` | TLS proxy listen address |
| `TLS_CERT_FILE` | `certs/server.crt` | Certificate path |
| `TLS_KEY_FILE` | `certs/server.key` | Private key path |

---

## TLS certificates

Auto-created in `certs/` on first start if files are missing.

Manual generation:

```powershell
.\scripts\generate-certs.ps1
```

```bash
sh scripts/generate-certs.sh
# or: make certs-sh
```

Cert files are gitignored.

---

## Tests

```bash
go test ./...
go test -race ./...
make test
make test-race
```

---

## Docker

```bash
docker compose up --build
```

Ports: `8080` (backend), `8443` (HTTPS proxy).  
Volume `./certs` is writable so certs can be generated on first run.

```bash
curl http://localhost:8080/health
curl -k https://localhost:8443/health
```

---

## Project layout

| Path | Role |
|------|------|
| `main.go` | Wiring, graceful shutdown |
| `internal/echo` | Echo API (`/echo`, `/health`) |
| `internal/proxy` | Reverse proxy, forwarding headers |
| `internal/certs` | Self-signed TLS generation |
| `internal/config` | Env-based configuration |

---

## Make targets

| Target | Action |
|--------|--------|
| `make run` | Run locally |
| `make test` | Unit tests |
| `make test-race` | Tests with race detector |
| `make build` | Build binary to `bin/` |
| `make docker` | `docker compose up --build` |
| `make certs-sh` | Generate certs (shell) |

---

## Graceful shutdown

1. Start: `go run .`
2. Send a long request: `curl "http://localhost:8080/echo?delay=10"`
3. Press `Ctrl+C`

Both servers stop accepting new connections and wait up to 10s for in-flight requests.

---

## License

MIT — see [LICENSE](LICENSE).
