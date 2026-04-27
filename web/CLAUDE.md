# Package: web

## Purpose
Embedded Go HTTP server that:
1. Serves the single-page dashboard (web/static/)
2. Exposes a JSON API: GET /api/events?window=4&lb=my-lb  (window = integer hours 1–24, default 1)
3. Exposes a CSV download: GET /api/export?window=4&lb=my-lb

## Implementation Notes
- Use Go 1.16+ embed directive to embed static/ into the binary
- server.go wires up routes using net/http ServeMux
- handlers.go calls internal/api.FetchEvents then either returns JSON or
  passes data to internal/export.WriteCSV for the export endpoint
- The /api/export handler must set headers:
    Content-Type: text/csv
    Content-Disposition: attachment; filename="sec_events_<timestamp>.csv"

## UI Design (current — updated 2026-04-27)
- Dark theme — F5 red (#E4003A) accent, dark charcoal backgrounds
- Top stats bar: Total Events | Blocked | Allowed | Top Event Type
- Chart 1: Timeline bar chart — 5-min buckets (Chart.js 4.4.1, no SRI hash)
- Chart 2: Doughnut — distribution by `sec_event_type` (not attack_type)
- Table: 10 columns, paginated, sortable, click-to-expand detail panel
- Controls: window slider (1–24 hours, label shows "Last N hour(s)"), LB input, Refresh, Export CSV

## Implementation Status: COMPLETE

### server.go
- `//go:embed static` on package-level `var staticFiles embed.FS`
- `fs.Sub(staticFiles, "static")` strips prefix so `/app.js` → `static/app.js`
- Routes: `/api/events`, `/api/export`, `/api/config`, `/` (static FileServer)
- `Start(port)` prints "Dashboard ready → http://localhost:<port>" to stderr then blocks
- **Build note**: static files are embedded at compile time. Must `go build` after any static file change — `go run` may serve stale embeds on some platforms.

### handlers.go
- `queryParams(r)` helper: returns hours int (default 1, clamped 1–24) and lb from query string; parses `?window=` as integer
- `clientForRequest()`: resolves API key per-request (X-Api-Key header → env var → 401)
- `eventsHandler`: nil events slice coerced to `[]api.SecurityEvent{}` before JSON encode
- `exportHandler`: sets `Content-Type: text/csv` and `Content-Disposition: attachment; filename="sec_events_<UTC timestamp>.csv"`
- Both handlers return HTTP 405 on non-GET; HTTP 502 on FetchEvents error

### static/index.html
- Chart.js 4.4.1 from cdnjs — NO integrity/crossorigin/referrerpolicy (hash was invalid, blocked both charts)
- Table thead: Time | Country | City | Src IP | Method | Rsp Code | Event Type | Action | Domain | Path
- data-col attributes match SecurityEvent JSON field names for sort
- IDs: `loading`, `error-banner`, `error-text`, `window-slider`, `window-label`, `lb-input`,
  `stat-total`, `stat-blocked`, `stat-allowed`, `stat-top-attack`,
  `timeline-chart`, `doughnut-chart`, `events-table`, `events-tbody`, `table-info`, `pagination`

### static/app.js
- State: `allEvents`, `filtered`, `currentPage`, `PAGE_SIZE=15`, `sortCol='time'`,
  `sortDir='desc'`, `currentWindow=1` (integer hours), `expandedIdx=null`
- Charts created once in `initCharts()`, updated in-place — never recreated
- Doughnut buckets by `e.sec_event_type`; stats blocked/allowed use `e.action`
- `renderTable()`: 10-column rows with `▶`/`▼` expand indicator in Time cell
- `toggleRow(idx)`: sets `expandedIdx`; re-render inserts detail `<tr>` after selected row
- `buildDetailPanel(e)`: 4-section 2-column grid — Src | Request | Detection | Signatures
  - Signatures section spans full width; each signature rendered as `.sig-card`
  - `parseJsonField(val)`: handles `undefined`, `null`, raw JS array, and double-encoded JSON strings
- Sort/page change/new fetch resets `expandedIdx` to null
- `escHtml()` on all data rendered into innerHTML (XSS safe)

### static/style.css
- CSS variables on `:root` — change `--accent` to re-theme
- `.data-row` — cursor pointer; `.row-expanded` — highlighted background
- `.row-block` / `.row-allow` — red/green left border on action rows
- `.badge-block/.badge-allow/.badge-other` — WAF action chips
- `.detail-panel` — 2-col CSS grid inside expanded `<tr>`
- `.detail-section` — individual card (Source/Request/Detection/Signatures)
- `.detail-section.detail-signatures` — spans both columns (grid-column: 1/-1)
- `.detail-grid` — label/value pair grid (max-content + 1fr)
- `.sig-card` — individual signature card within Signatures section
- `.window-slider-row`, `.window-slider` (accent-color red), `.window-label` — slider UI replacing old toggle buttons
- `.key-row`, `.key-status.key-ok/.key-missing`, `.setup-prompt` — connection UI

## UI Settings Panel
- Sidebar "Connection": API Key (password + eye toggle), Namespace input
- Key in sessionStorage (SS_KEY), namespace in SS_NAMESPACE, lb in SS_LB
- On load: `loadServerConfig()` seeds namespace; `restoreSettings()` applies session values (session wins)
- `buildHeaders(apiKey, namespace)` → `{X-Api-Key, X-Namespace}` on every fetch
- `exportCSV()` uses fetch + blob download — key in header, never in URL
- 401 → JSON `{error: "..."}` shown in error banner
- `showSetupPrompt(bool)` toggles `#setup-prompt` / `#dashboard`
