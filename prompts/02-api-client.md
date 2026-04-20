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
