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

## Implementation Status: COMPLETE (Prompt 3)

### server.go
- `//go:embed static` on package-level `var staticFiles embed.FS`
- `fs.Sub(staticFiles, "static")` strips prefix so `/app.js` → `static/app.js`
- Routes: `/api/events`, `/api/export`, `/` (static FileServer)
- `Start(port)` prints "Dashboard ready → http://localhost:<port>" to stderr then blocks

### handlers.go
- `queryParams(r)` helper: returns window (default "1h") and lb from query string
- `eventsHandler`: nil events slice coerced to `[]api.SecurityEvent{}` before JSON encode
- `exportHandler`: sets `Content-Type: text/csv` and `Content-Disposition: attachment; filename="sec_events_<UTC timestamp>.csv"` before calling `export.WriteCSV`
- Both handlers return HTTP 405 on non-GET; HTTP 502 on FetchEvents error

### static/index.html
- Loads Chart.js 4.4.1 from cdnjs with SRI integrity hash
- Layout: sidebar (controls) + main (stats → charts → table)
- IDs used by app.js: `loading`, `error-banner`, `error-text`, `btn-1h`, `btn-24h`,
  `lb-input`, `stat-total`, `stat-blocked`, `stat-allowed`, `stat-top-attack`,
  `timeline-chart`, `doughnut-chart`, `events-table`, `events-tbody`, `table-info`, `pagination`

### static/app.js
- Module-level state: `allEvents`, `filtered`, `currentPage`, `PAGE_SIZE=15`, `sortCol='time'`, `sortDir='desc'`, `currentWindow='1h'`
- `timelineChart` and `doughnutChart` created once in `initCharts()`, updated in-place — never recreated
- Timeline buckets: 5-min intervals keyed by Unix-ms, sorted numerically
- `sortBy(col)` toggles asc/desc on re-click; `applySort()` mutates `filtered`
- `escHtml()` used on all event field data before innerHTML injection (XSS safe)
- Loading spinner via `.hidden` CSS class toggle; error banner auto-dismissed with `dismissError()`

### static/style.css
- CSS variables on `:root` for all colours — change `--accent` to re-theme
- `.row-block` = left red border; `.row-allow` = left green border
- `.badge-block/.badge-allow/.badge-other` for WAF action chips
- `.sev-critical/.sev-high/.sev-medium/.sev-low` for severity colouring
- `.key-row` flex container for password input + show/hide button
- `.key-status.key-ok/.key-missing` — green/grey API key status indicator
- `.setup-prompt` — centered onboarding screen shown when no API key is set

## UI Settings Panel (post-Prompt-5)
- Sidebar "Connection" section: API Key (password field + eye toggle), Namespace input
- API key stored in sessionStorage (SS_KEY), namespace in SS_NAMESPACE, lb in SS_LB
- On load: restoreSettings() reads sessionStorage; if key present, auto-fetches; else shows setup prompt
- buildHeaders(apiKey, namespace) returns {X-Api-Key, X-Namespace} sent with every fetch
- exportCSV() uses fetch() + blob download so key goes in header, not URL
- 401 response: parsed as JSON {error: "..."} and shown in error banner
- showSetupPrompt(bool) toggles #setup-prompt / #dashboard visibility
