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
- Time window: integer hours 1–24 — passed as CLI flag `--window <int>` (default 1)
  and as query param `?window=<int>` in the web API (default 1, clamped 1–24)

## F5 XC API
- Base URL pattern: https://{tenant}.console.ves.volterra.io/api/data/namespaces/{namespace}/app_security/events
- Auth header: Authorization: APIToken {api_key}
- POST body: start_time, end_time (RFC3339), optional query filter using `vh_name` field
- Query filter format: `{vh_name="ves-io-http-loadbalancer-{lb-name}"}` — NOT virtual_host
- vh_name pattern confirmed 2026-04-27: `ves-io-http-loadbalancer-{lbname}` (not `ves-io-{namespace}-{lbname}`)
- Response `events` array contains JSON-encoded strings (double-encoded) — two-pass unmarshal required
- See docs/api-reference.md for exact request/response shapes and field type quirks

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
# CLI mode (prints JSON) — --window is hours 1–24; use full ves-io-http-loadbalancer-{lb-name} for --lb
F5XC_API_KEY=xxx /home/coder/go/bin/go run ./cmd/f5xc-sec --window 4 --namespace s-iannetta --lb ves-io-http-loadbalancer-webuiaz

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
- ALL 5 PROMPTS COMPLETE + UI settings + namespace switching + live API field fixes + detail panel UI + 1–24h slider
- `go build ./...`, `go test ./...` (20 tests), `go vet ./...` all pass
- Live curl confirmed 2026-04-27: 58 events (malicious_user_sec_event + waf_sec_event), zero errors
- Web server starts without env key: `./bin/f5xc-sec --serve --port 8080` → paste key in browser
- Events table: 10 columns with click-to-expand detail panel (Src/Request/Detection/Signatures)
- Time window: range slider 1–24 hours (replaces 1h/24h toggle buttons); label shows "Last N hour(s)"
- GET /api/config seeds namespace from server config; user can override in browser
- CLI/export still require F5XC_API_KEY env var
- Windows build: always use `go build` then run the binary — `go run` may serve stale embedded files

## Confirmed Field Types (live API 2026-04-27 — do not change without re-testing)
- `latitude`, `longitude` → `string` (quoted string despite looking numeric)
- `start_time`, `end_time` (event payload) → `int64` (Unix epoch seconds)
- Score fields (suspicion_score, waf_suspicion_score, etc., 8 fields) → `float64`
- `feature_score` → `string` exception (API sends as JSON-encoded string `"{}"`)
- Count fields (req_count, waf_sec_event_count, etc., 8 fields) → `int`
- `apiep_anomaly` → `int`
- Per-request fields (rsp_code, req_size, rsp_size, upstream_rsp_code) → `string` (unconfirmed int; string is safe default for this API)
- `signatures`, `req_risk_reasons` → `json.RawMessage` (arrays in live API; may be double-encoded strings)
- Chart.js loaded without SRI hash (hash was invalid, blocked both charts)
