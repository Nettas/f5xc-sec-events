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

## Known Field Type Quirks (live API)
- `latitude`, `longitude` — sent as JSON strings, not numbers → typed as string in struct
- `start_time`, `end_time` in SecurityEvent — sent as Unix epoch int64, not RFC3339 strings
- Request body `start_time`/`end_time` (eventsRequest) — still RFC3339 strings as required by the API

## Implementation Status: COMPLETE
- models.go: SecurityEvent (50+ real fields), EventsResponse{RawEvents []string}, eventsRequest
- client.go: Client{tenant, apiKey, httpClient, baseURL}; NewClient(); WithTimeout(); WithAPIKey(); WithBaseURL()
  - baseURL field is a test seam — when non-empty it overrides the real F5 XC URL
  - Do NOT remove baseURL; client_test.go depends on it via newTestClient()
- events.go: FetchEvents(ctx, namespace, lbName, window) — two-pass unmarshal, debug logging to stderr
  - lbName="" → no query filter; lbName set → query=`{vh_name="%s"}`
- client_test.go: 5 tests, all passing (1h window, 24h window, auth header, non-200, invalid window)
