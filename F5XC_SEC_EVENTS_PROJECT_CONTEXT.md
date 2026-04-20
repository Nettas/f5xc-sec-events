F5 XC Security Events Tool — Project Context & Session Notes
> **Purpose of this file**: Drop this into Claude Code at the start of any session to restore full project context without re-reading the chat. Last updated: 2026-04-14.
---
Project Summary
Building a Go CLI + embedded web dashboard that:
Authenticates to F5 Distributed Cloud (tenant: `f5-sa`) using an API key
Pulls WAF/security events from a specific namespace and HTTP Load Balancer
Supports two time windows: last 1 hour or last 24 hours
Visualizes events in a browser-based dashboard (dark theme, Chart.js)
Exports events to CSV for Excel use
---
Key Decisions Made
Decision	Choice	Reason
Language	Go 1.22+	User requirement
Coding tool	Claude Code (not Claude chat)	User wants agentic coding
Repo pattern	Custom purpose-built repo	agency-agents and claude-cookbooks don't fit Go CLI/API tooling
CLAUDE.md strategy	Root + per-package	Claude Code reads these automatically; keeps context scoped
UI	Vanilla JS + Chart.js embedded in Go binary	No build step; single binary deployment
Auth	`APIToken` header via environment variable	Matches F5 XC API spec; no hardcoded secrets
Time windows	`--window 1h|24h` CLI flag + `?window=` query param	Consistent across CLI and web modes
CSV export	Both CLI (`--export` flag) and web (`/api/export` endpoint)	Flexibility for automation and browser use
Rejected repos	`agency-agents` (marketing/persona prompts) and `claude-cookbooks` (Jupyter/Anthropic API)	Neither supports Go CLI tooling patterns
---
F5 XC API Details
Tenant: `f5-sa`
Base URL: `https://f5-sa.console.ves.volterra.io/api/data/namespaces/{namespace}/app_security/events`
Namespace: `s-iannetta`
Auth Header: `Authorization: APIToken <F5XC_API_KEY>`
Method: POST
API Docs: https://docs.cloud.f5.com/docs-v2/api
Request Body (JSON)
```json
{
  "start_time": "<RFC3339>",
  "end_time":   "<RFC3339>",
  "namespace":  "s-iannetta",
  "query": "{virtual_host=\"my-lb\"}"
}
```
Response Shape (approximate — verify against live API)
```json
{
  "events": [
    {
      "time": "...",
      "src_ip": "...",
      "req_path": "...",
      "method": "...",
      "response_code": 403,
      "req_id": "...",
      "waf_action": "BLOCK|ALLOW",
      "attack_type": "...",
      "severity": "...",
      "virtual_host": "..."
    }
  ]
}
```
---
Repo Structure
```
f5xc-sec-events/
├── CLAUDE.md                        ← Root agent instructions
├── README.md
├── go.mod
├── go.sum
├── .env.example
├── .gitignore
│
├── cmd/
│   └── f5xc-sec/
│       ├── CLAUDE.md
│       └── main.go
│
├── internal/
│   ├── api/
│   │   ├── CLAUDE.md
│   │   ├── client.go
│   │   ├── client_test.go
│   │   ├── events.go
│   │   └── models.go
│   ├── export/
│   │   ├── CLAUDE.md
│   │   └── csv.go
│   └── config/
│       ├── CLAUDE.md
│       └── config.go
│
├── web/
│   ├── CLAUDE.md
│   ├── server.go
│   ├── handlers.go
│   └── static/
│       ├── index.html
│       ├── app.js
│       └── style.css
│
├── docs/
│   ├── api-reference.md
│   └── architecture.md
│
└── prompts/
    ├── 01-scaffold.md
    ├── 02-api-client.md
    ├── 03-ui-dashboard.md
    ├── 04-csv-export.md
    └── 05-testing-hardening.md
```
---
CLAUDE.md Files (Full Content)
Root `CLAUDE.md`
```markdown
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
- web/              Embedded web server and static 2dashboard UI
- prompts/          Sequenced Claude Code prompts — run these in order

## Running the tool
\`\`\`bash
# CLI mode (prints JSON)
F5XC_API_KEY=xxx go run ./cmd/f5xc-sec --window 1h --namespace s-iannetta --lb my-lb

# Web server mode
F5XC_API_KEY=xxx go run ./cmd/f5xc-sec --serve --port 8080
\`\`\`
```
`internal/api/CLAUDE.md`
```markdown
# Package: internal/api

## Purpose
HTTP client for the F5 Distributed Cloud app_security events endpoint.

## Key Files
- client.go   — creates an *http.Client with timeout; attaches APIToken auth header
- models.go   — Go structs that map to the F5 XC JSON response (use json tags)
- events.go   — FetchEvents(ctx, window, namespace, lbName) → []SecurityEvent, error

## F5 XC Events API Details
Endpoint: POST https://{tenant}.console.ves.volterra.io/api/data/namespaces/{namespace}/app_security/events

Request body (JSON):
{
  "start_time": "<RFC3339>",
  "end_time":   "<RFC3339>",
  "namespace":  "s-iannetta",
  "query": "{virtual_host=\"my-lb\"}"
}

Auth: Header → Authorization: APIToken <F5XC_API_KEY>

## Notes
- Compute start_time as time.Now().Add(-1*time.Hour) or -24*time.Hour based on window flag
- Return a typed error if HTTP status != 200
- Use encoding/json for marshal/unmarshal
```
`web/CLAUDE.md`
```markdown
# Package: web

## Purpose
Embedded Go HTTP server that:
1. Serves the single-page dashboard (web/static/)
2. Exposes a JSON API: GET /api/events?window=1h&lb=my-lb
3. Exposes a CSV download: GET /api/export?window=1h&lb=my-lb

## Implementation Notes
- Use Go 1.16+ embed directive to embed static/ into the binary
- server.go wires up routes using net/http ServeMux
- handlers.go calls internal/api.FetchEvents then either returns JSON or
  passes data to internal/export.WriteCSV for the export endpoint
- The /api/export handler must set headers:
    Content-Type: text/csv
    Content-Disposition: attachment; filename="sec_events_<timestamp>.csv"

## UI Design Goals (web/static/)
- Dark theme dashboard — F5 red (#E4003A) as accent, dark charcoal backgrounds
- Top stats bar: Total Events | Blocked | Allowed | Top Attack Type
- Chart 1: Timeline bar chart of events by 5-min bucket (Chart.js)
- Chart 2: Doughnut chart of attack type distribution
- Table: Paginated sortable event table (src_ip, path, method, action, attack_type, time)
- Controls: window toggle (1h / 24h), LB name input, Refresh button, Export CSV button
- Use Chart.js from CDN (cdnjs.cloudflare.com)
- No frameworks. Vanilla JS fetch() to /api/events
```
`internal/export/CLAUDE.md`
```markdown
# Package: internal/export

## Purpose
Serialize a []api.SecurityEvent slice to CSV format.

## Requirements
- Use encoding/csv (stdlib only, no third-party libs)
- CSV columns (in order):
    time, src_ip, method, req_path, response_code, waf_action,
    attack_type, severity, virtual_host, req_id
- WriteCSV(w io.Writer, events []api.SecurityEvent) error
- Must flush the csv.Writer before returning
- Dates in CSV should be human-readable (RFC3339 format)
```
---
Sequenced Claude Code Prompts
Run these in order inside `claude` (Claude Code CLI). Each prompt is also saved under `prompts/` in the repo.
Prompt 1 — Project Scaffold (`prompts/01-scaffold.md`)
```
Read CLAUDE.md first. Then:

1. Initialize a Go module named `github.com/Nettas/f5xc-sec-events`
2. Create the full directory tree from CLAUDE.md's Directory Guide
3. Create a .env.example with: F5XC_API_KEY=, F5XC_TENANT=f5-sa, F5XC_NAMESPACE=s-iannetta
4. Create .gitignore that ignores: .env, *.env, bin/, vendor/
5. Create a minimal go.mod
6. In cmd/f5xc-sec/main.go create a skeleton main() that:
   - Parses flags: --window (default "1h"), --namespace, --lb, --serve, --port (default 8080)
   - Loads config from internal/config
   - Prints "F5 XC Security Events Tool" and exits cleanly
7. Make sure `go build ./...` succeeds with no errors
```
Prompt 2 — F5 XC API Client (`prompts/02-api-client.md`)
```
Read CLAUDE.md and internal/api/CLAUDE.md first. Then:

1. Implement internal/api/models.go — define SecurityEvent struct and EventsResponse
   struct with proper json tags.

2. Implement internal/api/client.go:
   - Client struct holding tenant, apiKey, httpClient
   - NewClient(tenant, apiKey string) *Client
   - Set http.Client timeout to 30s

3. Implement internal/api/events.go:
   - FetchEvents(ctx context.Context, namespace, lbName, window string) ([]SecurityEvent, error)
   - Compute start_time / end_time from window string ("1h" or "24h")
   - Build POST request body per spec
   - Set Authorization: APIToken {apiKey} header
   - Parse response into EventsResponse and return the Events slice
   - Return descriptive wrapped errors on non-200 status

4. Write internal/api/client_test.go:
   - Use httptest.NewServer to mock the F5 XC API
   - Test 1h and 24h windows return correct start_time
   - Test non-200 returns an error

5. Run `go test ./internal/api/...` and fix any failures
```
Prompt 3 — Web Dashboard + API Server (`prompts/03-ui-dashboard.md`)
```
Read CLAUDE.md and web/CLAUDE.md first. Then:

1. Implement web/server.go:
   - NewServer(client *api.Client, cfg *config.Config) *Server
   - Start(port int) error using net/http
   - Use //go:embed static to embed static files
   - Register routes: GET /api/events, GET /api/export, GET / (static)

2. Implement web/handlers.go:
   - eventsHandler: reads ?window= and ?lb= query params, calls client.FetchEvents,
     returns JSON array of events
   - exportHandler: same but streams CSV using internal/export.WriteCSV with
     correct Content-Disposition header

3. Implement web/static/index.html, app.js, style.css per design goals in web/CLAUDE.md:
   - Dark theme, F5 red accent (#E4003A)
   - Stats bar, timeline chart, attack-type doughnut, paginated events table
   - Window toggle (1h / 24h), LB input, Refresh, Export CSV button
   - Load Chart.js from cdnjs CDN

4. Wire --serve flag in cmd/f5xc-sec/main.go to start the web server

5. Test manually: `go run ./cmd/f5xc-sec --serve --port 8080`
```
Prompt 4 — CSV Export (`prompts/04-csv-export.md`)
```
Read internal/export/CLAUDE.md first. Then:

1. Implement internal/export/csv.go per spec (columns, order, flush)
2. Integrate into web/handlers.go exportHandler
3. Also support CLI export mode:
   - When `--export` flag is set, run FetchEvents and pipe through WriteCSV to stdout
   - Example: `go run ./cmd/f5xc-sec --window 24h --lb my-lb --export > events.csv`
4. Write a unit test for WriteCSV with 3 mock SecurityEvent records,
   verify CSV header row and data rows
```
Prompt 5 — Testing, Error Handling & Polish (`prompts/05-testing-hardening.md`)
```
1. Add integration-style tests for web handlers using httptest.NewRecorder
2. Ensure all API errors surface as user-friendly messages in both CLI and UI
3. Add a --timeout flag (seconds) to the CLI that overrides the default 30s
4. In the UI: show a loading spinner while fetching, an error banner on failure
5. Add a README.md that covers:
   - Prerequisites (Go 1.22+)
   - How to get an F5 XC API key
   - Running CLI mode (1h and 24h examples)
   - Running web server mode
   - How to export CSV
6. Run `go vet ./...` and `go test ./...` — fix all issues
```
---
Environment Variables
```bash
F5XC_API_KEY=<your-api-token>
F5XC_TENANT=f5-sa
F5XC_NAMESPACE=s-iannetta
```
---
How to Start Claude Code on This Project
```bash
# 1. Create the repo
mkdir f5xc-sec-events && cd f5xc-sec-events
git init

# 2. Create all CLAUDE.md files (content above) and prompts/ folder

# 3. Launch Claude Code
claude

# 4. Run prompts in order — paste content from each prompts/0X-*.md file
#    OR use: /read prompts/01-scaffold.md  (if supported in your claude version)
```
---
Status / Progress Tracker
[x] Prompt 1 — Scaffold complete (2026-04-16)
[x] Prompt 2 — API client complete + tests passing (2026-04-16)
[x] Prompt 3 — Web dashboard + server complete (2026-04-16)
[x] Prompt 4 — CSV export working (CLI + web) (2026-04-16)
[x] Prompt 5 — Tests passing, `go vet` clean (2026-04-16)
---
Implementation Notes (non-obvious decisions)

Prompt 1 — Scaffold
- Go is NOT in PATH on this machine. Installed to /home/coder/go/bin/go.
  Always invoke as: /home/coder/go/bin/go build ./...
  (GOPATH=GOROOT warning is harmless — module mode ignores it)
- Module path: github.com/Nettas/f5xc-sec-events (replace with real org before publishing)
- internal/config/config.go: Load() returns error if F5XC_API_KEY is empty;
  F5XC_TENANT defaults to "f5-sa", F5XC_NAMESPACE defaults to "s-iannetta"

Prompt 2 — API Client
- Client has a `baseURL string` field (zero value = use real F5 XC URL).
  Tests inject the httptest.Server URL via newTestClient() in client_test.go.
  This is a test seam — don't remove it.
- Client.WithTimeout(d) returns a new Client copy; wired in main.go via --timeout flag.
- eventsRequest is unexported — only used internally in events.go.
- FetchEvents query string uses %q formatting — verify against live API.
- If lbName is empty string, query field is omitted (no filter).
- All 5 tests pass.

Prompt 3 — Web Dashboard
- internal/export/csv.go: WriteCSV implemented here (Prompt 4 will add CLI mode + tests).
- web/server.go: embed directive `//go:embed static` on package-level var staticFiles.
  fs.Sub strips the "static/" prefix so URLs map directly (e.g. /app.js → static/app.js).
- web/handlers.go: queryParams() helper extracts window/lb; nil events coerced to [].
  exportHandler sets Content-Disposition: attachment; filename="sec_events_<timestamp>.csv"
- web/static/app.js: all state is module-level vars (allEvents, filtered, currentPage,
  sortCol, sortDir, currentWindow). Two Chart.js instances (timelineChart, doughnutChart)
  are created once in initCharts() and updated in-place — never recreated.
- Timeline uses 5-min buckets; bucket keys are Unix-ms integers, sorted numerically.
- Table PAGE_SIZE=15; sortBy() toggles asc/desc on re-click of same column.
- escHtml() used on all user-originated data rendered into innerHTML.
- cmd/f5xc-sec/main.go: --serve wired; --export still stubbed (_ = export); CLI mode
  calls FetchEvents + json.Encoder to stdout; --timeout wired via WithTimeout.
- go build, go test, go vet all clean after Prompt 3.

Prompt 4 — CSV Export CLI
- Flag variable renamed `doExport` (was `export`) to avoid shadowing the `export` package import.
- `--export` block placed before the JSON default block, both sharing the same FetchEvents call.
- csv_test.go: 4 tests — HeaderRow, RowCount, DataRows (all 10 columns asserted per row), EmptySlice.
- go build, go test (9 tests total), go vet all clean after Prompt 4.

Prompt 5 — Testing, Hardening & README
- Added WithBaseURL(url string) *Client to api/client.go — exported test seam for web tests.
- web/handlers_test.go: 9 tests in package web (white-box), using newTestServer() helper that
  spins up httptest.NewServer as mock F5 backend and injects via WithBaseURL.
  Tests: ReturnsJSONArray, DefaultWindow, MethodNotAllowed×2, UpstreamError→502,
  EmptyEventsIsArray, ContentTypeCSV, ContentDisposition, CSVHasHeaderAndRows.
- README.md created: prerequisites, API key setup steps (F5 XC console walkthrough),
  build, web dashboard, CLI mode, CSV export, all flags table, env vars table, troubleshooting.
- ALL PROMPTS COMPLETE. 18/18 tests pass. go build + go vet clean.
---
Open Questions / Next Steps
- Verify exact API request/response shape — models.go fields should be validated
  against a live curl before finalizing. The virtual_host query filter syntax in
  particular needs live confirmation (%q quoting may need adjustment).
- Pagination — F5 XC events API may paginate. Check response for continuation tokens;
  add a pagination loop in events.go if needed.
- Go module path — replace github.com/Nettas/f5xc-sec-events with actual org/repo.
---
UI Settings Feature (post-Prompt-5) — 2026-04-16
- Added Connection section to sidebar: API key (password input + show/hide toggle),
  namespace input (pre-filled "s-iannetta").
- API key stored in sessionStorage (cleared on tab close — never persists across sessions).
- Each fetch sends X-Api-Key and X-Namespace request headers; key never appears in URL.
- web/handlers.go: clientForRequest() resolves client per-request:
    1. X-Api-Key header (UI) → client.WithAPIKey(key)
    2. s.cfg.APIKey (env var) → s.client
    3. Neither → 401 JSON error
- internal/api/client.go: added WithAPIKey(key string) *Client
- internal/config/config.go: added Defaults() — returns config with empty APIKey,
  used by --serve mode when F5XC_API_KEY env var is not set.
- cmd/f5xc-sec/main.go: --serve now starts without API key (logs warning, uses Defaults()).
  CLI and export modes still require F5XC_API_KEY env var and exit on error.
- No-key state: dashboard shows setup prompt with instructions instead of blank charts.
- go build + go test (18 tests) + go vet all clean.

Namespace Switching (post-UI-settings) — 2026-04-16
- Added GET /api/config endpoint (configHandler) — returns {namespace, tenant}, never api_key.
- Registered at /api/config in server.go registerRoutes().
- app.js: loadServerConfig() called on DOMContentLoaded before restoreSettings().
  Seeds namespace input from server config ONLY if sessionStorage has no saved value.
  sessionStorage always wins (user preference persists across refreshes within the same tab).
- Removed hardcoded value="s-iannetta" from index.html namespace input.
  Placeholder now reads "Namespace (e.g. s-iannetta)".
- Removed || 's-iannetta' fallback from getNamespace() — blank field sends blank header,
  server falls back to its own cfg.Namespace (which came from F5XC_NAMESPACE env var or default).
- Two new tests: TestConfigHandler_ReturnsNamespaceAndTenant, TestConfigHandler_MethodNotAllowed.
  Also asserts api_key is NOT present in response.
- Total tests: 20. go build + go test + go vet all clean.

---
Live API Triage & Field Fixes — 2026-04-20

Root cause of zero events was confirmed against live API (namespace: s-iannetta):

1. Query filter field — `virtual_host` does not exist. Correct field: `vh_name`
   - Format: `{vh_name="%s"}` using %s (not %q — no Go escaping)
   - Value pattern: `ves-io-{namespace}-{lb-name}` e.g. `ves-io-s-iannetta-webuiaz`

2. Wire format — `events` array contains JSON-encoded strings, not objects.
   Two-pass unmarshal added to events.go:
   - Pass 1: Unmarshal response → EventsResponse{RawEvents []string}
   - Pass 2: Loop, unmarshal each string → SecurityEvent

3. SecurityEvent struct replaced entirely with real API fields (50+ fields).
   Old fields removed: req_path, method, response_code, waf_action, attack_type,
   severity, virtual_host, req_id — none exist in live API.

4. Field type quirks confirmed against live data:
   - latitude, longitude → JSON string (not number) → Go string
   - start_time, end_time in event payload → Unix epoch int64 (not RFC3339 string)
   - Request body start_time/end_time remain RFC3339 strings (API requirement)

5. CSV columns updated to 14 real fields:
   time, src_ip, country, city, vh_name, app_type, threat_level, suspicion_score,
   waf_sec_event_count, req_count, waf_suspicion_score, summary_msg, namespace, tenant

6. Debug logging added to events.go (temporary):
   - [DEBUG] request body — printed after marshal, before HTTP call
   - [DEBUG] response status + body — printed after response read

7. handlers_test.go mock updated to marshalRawEvents() helper — builds mock response
   in the same double-encoded format as the real API.

All 20 tests pass. go build + go vet clean.

---
Score Field Type Fix — 2026-04-20 (follow-up)

All nine float64 score fields in SecurityEvent caused 502 unmarshal errors because the live API
sends them as JSON strings, not numbers. Fixed by changing all to Go `string`:

  suspicion_score, waf_suspicion_score, bot_defense_suspicion_score,
  behavior_anomaly_score, feature_score, ip_reputation_suspicion_score,
  forbidden_access_suspicion_score, failed_login_suspicion_score, rate_limit_suspicion_score

csv.go updated: score fields written directly (no %g formatting needed).
Test mock events updated to use string literals for these fields.
All 20 tests pass. go build + go vet clean.

OUTSTANDING RISK: int count fields (req_count, waf_sec_event_count, err_count, etc.) are
still typed as Go `int`. If a 502 "cannot unmarshal string into int" appears, change those
to `string` in models.go using the same pattern.

Last updated: 2026-04-20. ALL PROMPTS COMPLETE + UI settings + namespace switching + live API fixes.