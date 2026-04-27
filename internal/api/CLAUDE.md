# Package: internal/api

## Purpose
HTTP client for the F5 Distributed Cloud app_security events endpoint.

## Key Files
- client.go   — creates an *http.Client with timeout; attaches APIToken auth header
- models.go   — Go structs that map to the F5 XC JSON response (use json tags)
- events.go   — FetchEvents(ctx, namespace, lbName string, hours int) → []SecurityEvent, error

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
- vh_name pattern confirmed 2026-04-27: `ves-io-http-loadbalancer-{lbname}`
  (earlier docs said `ves-io-{namespace}-{lbname}` — that was wrong)

## Field Types — CONFIRMED against live API 2026-04-27 (58-event curl test)

> Locked. Any change requires a live re-test — type mismatches cause 502 unmarshal errors.

### Confirmed fields (malicious_user_sec_event + waf_sec_event aggregates)

| Field(s) | JSON type | Go type | Notes |
|----------|-----------|---------|-------|
| `latitude`, `longitude` | string | `string` | Quoted string despite looking numeric |
| `start_time`, `end_time` (event payload) | number | `int64` | Unix epoch seconds |
| `start_time`, `end_time` (request body) | string | `string` | RFC3339 — required by API |
| `suspicion_score`, `waf_suspicion_score`, `bot_defense_suspicion_score`, `behavior_anomaly_score`, `ip_reputation_suspicion_score`, `forbidden_access_suspicion_score`, `failed_login_suspicion_score`, `rate_limit_suspicion_score` | float | `float64` | All score fields are JSON floats |
| `feature_score` | string | `string` | Exception — API sends `"{}"` (JSON-encoded string) |
| `req_count`, `waf_sec_event_count`, `bot_defense_sec_event_count`, `err_count`, `failed_login_count`, `forbidden_access_count`, `page_not_found_count`, `rate_limiting_count` | number | `int` | JSON integers |
| `apiep_anomaly` | number | `int` | JSON integer (0 or 1) |
| `policy_hits` | object/null | `json.RawMessage` | Variable shape |
| `timeseries_enabled` | bool | `bool` | |
| `incremental_activity_info`, `method_counts`, `mitigation_activity_info` | string | not mapped | Double-encoded nested objects — silently ignored |

### Per-request fields (waf_sec_event individual hits — types unconfirmed, safe defaults used)

| Field(s) | Go type | Notes |
|----------|---------|-------|
| `method`, `action`, `domain`, `req_path`, `req_id`, `authority`, `api_endpoint`, `req_risk`, `browser_type`, `device_type`, `user_agent`, `tls_fingerprint`, `ja4_tls_fingerprint`, `src_site`, `src`, `req_params` | `string` | All string — safe for this API |
| `rsp_code`, `req_size`, `rsp_size`, `upstream_rsp_code` | `string` | Changed int→string 2026-04-27; API likely sends as strings, caused Windows 502 when int |
| `signatures` | `json.RawMessage` | Array of signature objects; may be double-encoded string |
| `req_risk_reasons` | `json.RawMessage` | Array of reason strings; may be double-encoded |

All per-request fields use `omitempty` — absent on aggregated event types, so they never reach the decoder for existing confirmed events.

## Implementation Status: COMPLETE
- models.go: SecurityEvent (confirmed fields + 22 new per-request/detail fields), EventsResponse, eventsRequest
  - Requires `import "encoding/json"` for json.RawMessage
  - Extra `map[string]json.RawMessage \`json:"-"\`` is a struct placeholder only — json:"-" means decoder never touches it; unknown API fields are silently discarded by default
- client.go: Client{tenant, apiKey, httpClient, baseURL}; NewClient(); WithTimeout(); WithAPIKey(); WithBaseURL()
  - baseURL is a test seam — do NOT remove; client_test.go depends on it via newTestClient()
- events.go: FetchEvents(ctx, namespace, lbName string, hours int) — two-pass unmarshal, debug logging to stderr
  - hours must be 1–24; returns error otherwise
  - startTime = now - hours*time.Hour
  - lbName="" → no filter; lbName set → query=`{vh_name="%s"}`
  - Plain json.Unmarshal — unknown fields silently ignored
- client_test.go: 5 tests passing (1h window, 24h window, auth header, non-200, invalid hours 0 and 25)
