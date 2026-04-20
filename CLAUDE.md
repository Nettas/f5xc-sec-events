# F5 XC Security Events Tool — Claude Code Agent Instructions

## Project Purpose
A Go CLI + web dashboard that pulls WAF/security events from F5 Distributed Cloud
(tenant: f5-sa) via the app_security events API, visualizes them in a browser UI,
and exports them to CSV.

## Tech Stack
- Language: Go 1.22+
- Web UI: Vanilla JS + Chart.js (no framework), served by Go's net/http
- Config: Environment variables + CLI flags (no hardcoded secrets)
- Testing: standard library testing package + httptest

## Conventions
- All API keys come from env vars: F5XC_API_KEY, F5XC_TENANT, F5XC_NAMESPACE
- Never hardcode credentials. Never commit .env files.
- Use Go modules. Package layout follows standard Go project layout.
- All HTTP calls must have configurable timeouts (default 30s).
- Errors must be wrapped with context using fmt.Errorf("...: %w", err)
- Log to stderr; structured output (JSON events) to stdout or HTTP response
- Time windows: "1h" = last 1 hour, "24h" = last 24 hours — passed as a CLI
  flag `--window 1h|24h` and as a query param `?window=1h` in the web API

## F5 XC API
- Base URL pattern: https://{tenant}.console.ves.volterra.io/api/data/namespaces/{namespace}/app_security/events
- Auth header: Authorization: APIToken {api_key}
- The API uses POST with a JSON body containing start_time, end_time, and optional
  filter for virtual_host (http-lb name)
- See docs/api-reference.md for exact request/response shapes

## Directory Guide
- cmd/f5xc-sec/     CLI entry point
- internal/api/     F5 XC HTTP client and models
- internal/export/  CSV export logic
- internal/config/  Config loading
- web/              Embedded web server and static dashboard UI
- prompts/          Sequenced Claude Code prompts — run these in order

## Environment
- Go is installed at `/home/coder/go/bin/go` — it is NOT in PATH
- Always invoke as: `/home/coder/go/bin/go build ./...` etc.
- The GOPATH=GOROOT warning that appears is harmless

## Running the tool
```bash
# CLI mode (prints JSON)
F5XC_API_KEY=xxx /home/coder/go/bin/go run ./cmd/f5xc-sec --window 1h --namespace s-iannetta --lb my-lb

# Web server mode
F5XC_API_KEY=xxx /home/coder/go/bin/go run ./cmd/f5xc-sec --serve --port 8080
```

## Build & Test
```bash
/home/coder/go/bin/go build ./...
/home/coder/go/bin/go test ./...
/home/coder/go/bin/go vet ./...
```

## Current Status
- ALL 5 PROMPTS COMPLETE + UI settings + namespace switching — `go build ./...`, `go test ./...` (20 tests), `go vet ./...` all pass
- Web server starts without env key: `./bin/f5xc-sec --serve --port 8080` → paste key in browser
- GET /api/config seeds namespace field from server config; user can override freely
- CLI/export still require F5XC_API_KEY env var
