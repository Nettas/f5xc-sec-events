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
