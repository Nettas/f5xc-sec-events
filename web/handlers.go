package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Nettas/f5xc-sec-events/internal/api"
	"github.com/Nettas/f5xc-sec-events/internal/export"
)

// clientForRequest resolves the API client and namespace for a single HTTP request.
//
// Priority order (highest wins):
//  1. X-Api-Key / X-Namespace request headers  (set by the dashboard UI)
//  2. Server-level config loaded from env vars  (set at startup)
//
// Returns nil if no API key is available from either source.
func (s *Server) clientForRequest(r *http.Request) (*api.Client, string) {
	apiKey    := r.Header.Get("X-Api-Key")
	namespace := r.Header.Get("X-Namespace")

	if namespace == "" {
		namespace = s.cfg.Namespace
	}

	switch {
	case apiKey != "":
		// UI-supplied key overrides whatever the server was started with.
		return s.client.WithAPIKey(apiKey), namespace
	case s.cfg.APIKey != "":
		// Fall back to the env-var key.
		return s.client, namespace
	default:
		return nil, namespace
	}
}

// eventsHandler handles GET /api/events?window=1h&lb=my-lb
// Returns a JSON array of SecurityEvent (never null).
func (s *Server) eventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	client, namespace := s.clientForRequest(r)
	if client == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, `{"error":"API key required — enter it in the dashboard"}`)
		return
	}

	hours, lb := queryParams(r)

	events, err := client.FetchEvents(r.Context(), namespace, lb, hours)
	if err != nil {
		http.Error(w, fmt.Sprintf("fetch events: %v", err), http.StatusBadGateway)
		return
	}
	if events == nil {
		events = []api.SecurityEvent{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		fmt.Fprintf(os.Stderr, "encode events: %v\n", err)
	}
}

// exportHandler handles GET /api/export?window=1h&lb=my-lb
// Streams a CSV file download.
func (s *Server) exportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	client, namespace := s.clientForRequest(r)
	if client == nil {
		http.Error(w, "API key required — enter it in the dashboard", http.StatusUnauthorized)
		return
	}

	hours, lb := queryParams(r)

	events, err := client.FetchEvents(r.Context(), namespace, lb, hours)
	if err != nil {
		http.Error(w, fmt.Sprintf("fetch events: %v", err), http.StatusBadGateway)
		return
	}
	if events == nil {
		events = []api.SecurityEvent{}
	}

	filename := fmt.Sprintf("sec_events_%s.csv", time.Now().UTC().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	if err := export.WriteCSV(w, events); err != nil {
		fmt.Fprintf(os.Stderr, "write CSV: %v\n", err)
	}
}

// configHandler handles GET /api/config
// Returns the server's default namespace and tenant so the UI can seed its fields.
// The API key is intentionally never returned.
func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"namespace": s.cfg.Namespace,
		"tenant":    s.cfg.Tenant,
	})
}

// queryParams extracts the time window (in hours, 1–24) and lb from the request.
func queryParams(r *http.Request) (hours int, lb string) {
	hours = 1
	if w := r.URL.Query().Get("window"); w != "" {
		if n, err := strconv.Atoi(w); err == nil && n >= 1 && n <= 24 {
			hours = n
		}
	}
	lb = r.URL.Query().Get("lb")
	return
}
