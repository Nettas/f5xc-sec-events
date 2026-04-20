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

## Implementation Status: COMPLETE
- models.go: SecurityEvent, EventsResponse (exported), eventsRequest (unexported)
- client.go: Client{tenant, apiKey, httpClient, baseURL}; NewClient(); WithTimeout()
  - baseURL field is a test seam — when non-empty it overrides the real F5 XC URL
  - Do NOT remove baseURL; client_test.go depends on it via newTestClient()
- events.go: FetchEvents(ctx, namespace, lbName, window) — computes RFC3339 times,
  builds POST body, sets Authorization: APIToken header, returns []SecurityEvent
  - lbName="" → no query filter; lbName set → query="{virtual_host=\"lbName\"}"
  - Virtual_host query syntax uses %q formatting — verify against live API
- client_test.go: 5 tests, all passing (1h window, 24h window, auth header, non-200, invalid window)
