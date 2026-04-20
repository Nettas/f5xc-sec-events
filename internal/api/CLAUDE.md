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

## Known Field Type Quirks (live API — all confirmed 2026-04-20)
- `latitude`, `longitude` — JSON string (not number) → Go `string`
- `start_time`, `end_time` in SecurityEvent payload — Unix epoch number → Go `int64`
- Request body `start_time`/`end_time` (eventsRequest) — RFC3339 string (API requirement)
- ALL score fields — JSON string (not number) → Go `string`:
    suspicion_score, waf_suspicion_score, bot_defense_suspicion_score,
    behavior_anomaly_score, feature_score, ip_reputation_suspicion_score,
    forbidden_access_suspicion_score, failed_login_suspicion_score, rate_limit_suspicion_score
- Count fields (`req_count`, `waf_sec_event_count`, `err_count`, etc.) — still `int`, unconfirmed
  If a 502 "cannot unmarshal string into int" appears, change those to string too.

## Implementation Status: COMPLETE
- models.go: SecurityEvent (50+ real fields), EventsResponse{RawEvents []string}, eventsRequest
- client.go: Client{tenant, apiKey, httpClient, baseURL}; NewClient(); WithTimeout(); WithAPIKey(); WithBaseURL()
  - baseURL field is a test seam — when non-empty it overrides the real F5 XC URL
  - Do NOT remove baseURL; client_test.go depends on it via newTestClient()
- events.go: FetchEvents(ctx, namespace, lbName, window) — two-pass unmarshal, debug logging to stderr
  - lbName="" → no query filter; lbName set → query=`{vh_name="%s"}`
- client_test.go: 5 tests, all passing (1h window, 24h window, auth header, non-200, invalid window)
