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
