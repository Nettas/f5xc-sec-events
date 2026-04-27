# Package: internal/api

## Purpose
HTTP client for the F5 Distributed Cloud app_security events endpoint.

## Key Files
- client.go   — creates an *http.Client with timeout; attaches APIToken auth header
- models.go   — Go structs that map to the F5 XC JSON response (use json tags)
- events.go   — FetchEvents(ctx, namespace, lbName, window) → []SecurityEvent, error

## F5 XC Events API Details
Endpoint: POST https://{tenant}.console.ves.volterra.io/api/data/namespaces/{namespace}/app_security/events

Request body (JSON):
{
  "start_time": "<RFC3339>",
  "end_time":   "<RFC3339>",
  "namespace":  "s-iannetta",
  "query": "{vh_name=\"ves-io-s-iannetta-my-lb\"}"
}

Auth: Header → Authorization: APIToken <F5XC_API_KEY>

## CRITICAL: Wire Format (confirmed against live API 2026-04-20)
- The `events` array contains JSON-encoded STRINGS, not objects:
    {"events": ["{\"time\":\"...\",\"src_ip\":\"...\"}", ...]}
- FetchEvents does a TWO-PASS unmarshal:
    1. Unmarshal response envelope → EventsResponse{RawEvents []string}
    2. Loop over RawEvents, unmarshal each string → SecurityEvent
- Query filter field is `vh_name`, NOT `virtual_host`
- Query format: backtick template `{vh_name="%s"}` — NOT %q (no Go escaping)
- vh_name pattern: `ves-io-{namespace}-{lb-name}` e.g. `ves-io-s-iannetta-webuiaz`

## Known Field Type Quirks (live API — all confirmed 2026-04-20; count/time fields updated 2026-04-27)
- `latitude`, `longitude` — JSON string (not number) → Go `string`
- `start_time`, `end_time` in SecurityEvent payload — changed to Go `string` (API sends as int but string is safer for unmarshal)
- Request body `start_time`/`end_time` (eventsRequest) — RFC3339 string (API requirement)
- ALL score fields — JSON string (not number) → Go `string`:
    suspicion_score, waf_suspicion_score, bot_defense_suspicion_score,
    behavior_anomaly_score, feature_score, ip_reputation_suspicion_score,
    forbidden_access_suspicion_score, failed_login_suspicion_score, rate_limit_suspicion_score
- ALL count fields — changed to Go `string` (API confirmed to send as JSON strings):
    req_count, waf_sec_event_count, bot_defense_sec_event_count, err_count,
    failed_login_count, forbidden_access_count, page_not_found_count, rate_limiting_count

## Additional Fields (added 2026-04-27)
- `policy_hits` — may be object/array/null → `json.RawMessage` with omitempty
- `timeseries_enabled` — bool → Go `bool` with omitempty
- `Extra map[string]json.RawMessage \`json:"-"\`` — struct placeholder; note json:"-" means
  the decoder does NOT populate it. Go's json.Unmarshal silently ignores unknown fields by
  default — no DisallowUnknownFields is used anywhere, so unknown API fields are safe.

## Implementation Status: COMPLETE
- models.go: SecurityEvent (50+ real fields + PolicyHits/TimeseriesEnabled/Extra), EventsResponse{RawEvents []string}, eventsRequest
  - Requires `import "encoding/json"` for json.RawMessage
- client.go: Client{tenant, apiKey, httpClient, baseURL}; NewClient(); WithTimeout(); WithAPIKey(); WithBaseURL()
  - baseURL field is a test seam — when non-empty it overrides the real F5 XC URL
  - Do NOT remove baseURL; client_test.go depends on it via newTestClient()
- events.go: FetchEvents(ctx, namespace, lbName, window) — two-pass unmarshal, debug logging to stderr
  - lbName="" → no query filter; lbName set → query=`{vh_name="%s"}`
  - Uses plain json.Unmarshal (not Decoder+DisallowUnknownFields) — unknown fields silently ignored
- client_test.go: 5 tests, all passing (1h window, 24h window, auth header, non-200, invalid window)
