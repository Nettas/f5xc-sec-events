package web

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Nettas/f5xc-sec-events/internal/api"
	"github.com/Nettas/f5xc-sec-events/internal/config"
)

func marshalRawEvents(events []api.SecurityEvent) string {
	raw := make([]string, len(events))
	for i, e := range events {
		b, _ := json.Marshal(e)
		raw[i] = string(b)
	}
	resp := struct {
		Events []string `json:"events"`
	}{Events: raw}
	b, _ := json.Marshal(resp)
	return string(b)
}

// ── Test helpers ──────────────────────────────────────────────────────────────

// mockF5Response is a minimal valid F5 XC API response with two events.
// Each event is JSON-encoded as a string, matching the real API wire format.
var mockF5Response = marshalRawEvents([]api.SecurityEvent{
	{
		Time: "2026-04-16T10:00:00Z", SrcIP: "1.2.3.4",
		Country: "US", City: "New York",
		VhName: "ves-io-test-ns-my-lb", AppType: "web",
		ThreatLevel: "high", SuspicionScore: 85.5,
		WafSecEventCount: 3, ReqCount: 10, WafSuspicionScore: 75.0,
		SummaryMsg: "SQL injection attempt", Namespace: "test-ns", Tenant: "f5-sa",
	},
	{
		Time: "2026-04-16T10:05:00Z", SrcIP: "5.6.7.8",
		Country: "GB", City: "London",
		VhName: "ves-io-test-ns-my-lb", AppType: "web",
		ThreatLevel: "low", SuspicionScore: 10.0,
		WafSecEventCount: 0, ReqCount: 5, WafSuspicionScore: 0,
		SummaryMsg: "", Namespace: "test-ns", Tenant: "f5-sa",
	},
})

// newTestServer starts a mock F5 XC backend and returns a wired web.Server and cleanup func.
func newTestServer(t *testing.T, f5Handler http.HandlerFunc) (*Server, func()) {
	t.Helper()
	mock := httptest.NewServer(f5Handler)
	client := api.NewClient("test-tenant", "test-key").WithBaseURL(mock.URL)
	cfg := &config.Config{Namespace: "test-ns", Tenant: "test-tenant", APIKey: "test-key"}
	srv := NewServer(client, cfg)
	return srv, mock.Close
}

// okF5Handler returns the standard two-event mock response.
func okF5Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(mockF5Response))
}

// ── configHandler tests ───────────────────────────────────────────────────────

func TestConfigHandler_ReturnsNamespaceAndTenant(t *testing.T) {
	srv, cleanup := newTestServer(t, okF5Handler)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	srv.configHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode config response: %v", err)
	}
	if body["namespace"] != "test-ns" {
		t.Errorf("namespace = %q, want test-ns", body["namespace"])
	}
	if body["tenant"] != "test-tenant" {
		t.Errorf("tenant = %q, want test-tenant", body["tenant"])
	}
	if _, hasKey := body["api_key"]; hasKey {
		t.Error("response must not include api_key")
	}
}

func TestConfigHandler_MethodNotAllowed(t *testing.T) {
	srv, cleanup := newTestServer(t, okF5Handler)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/config", nil)
	rec := httptest.NewRecorder()
	srv.configHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

// ── eventsHandler tests ───────────────────────────────────────────────────────

func TestEventsHandler_ReturnsJSONArray(t *testing.T) {
	srv, cleanup := newTestServer(t, okF5Handler)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/events?window=1h&lb=my-lb", nil)
	rec := httptest.NewRecorder()
	srv.eventsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var events []api.SecurityEvent
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("got %d events, want 2", len(events))
	}
	if events[0].SrcIP != "1.2.3.4" {
		t.Errorf("events[0].SrcIP = %q, want 1.2.3.4", events[0].SrcIP)
	}
	if events[1].VhName != "ves-io-test-ns-my-lb" {
		t.Errorf("events[1].VhName = %q, want ves-io-test-ns-my-lb", events[1].VhName)
	}
}

func TestEventsHandler_DefaultWindow(t *testing.T) {
	var capturedBody string
	srv, cleanup := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		capturedBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"events":[]}`))
	})
	defer cleanup()

	// No ?window= param — handler should default to "1h"
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	rec := httptest.NewRecorder()
	srv.eventsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	// The F5 API request body should contain a start_time (not empty) — confirms window was applied.
	if !strings.Contains(capturedBody, "start_time") {
		t.Errorf("F5 request body missing start_time; got: %s", capturedBody)
	}
}

func TestEventsHandler_MethodNotAllowed(t *testing.T) {
	srv, cleanup := newTestServer(t, okF5Handler)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/events", nil)
	rec := httptest.NewRecorder()
	srv.eventsHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestEventsHandler_UpstreamError(t *testing.T) {
	srv, cleanup := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/events?window=1h", nil)
	rec := httptest.NewRecorder()
	srv.eventsHandler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", rec.Code)
	}
}

func TestEventsHandler_EmptyEventsIsArray(t *testing.T) {
	srv, cleanup := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"events":[]}`))
	})
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/events?window=1h", nil)
	rec := httptest.NewRecorder()
	srv.eventsHandler(rec, req)

	body := strings.TrimSpace(rec.Body.String())
	if body != "[]" {
		t.Errorf("empty events body = %q, want []", body)
	}
}

// ── exportHandler tests ───────────────────────────────────────────────────────

func TestExportHandler_ContentTypeCSV(t *testing.T) {
	srv, cleanup := newTestServer(t, okF5Handler)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/export?window=1h", nil)
	rec := httptest.NewRecorder()
	srv.exportHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "text/csv" {
		t.Errorf("Content-Type = %q, want text/csv", ct)
	}
}

func TestExportHandler_ContentDisposition(t *testing.T) {
	srv, cleanup := newTestServer(t, okF5Handler)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/export?window=1h", nil)
	rec := httptest.NewRecorder()
	srv.exportHandler(rec, req)

	cd := rec.Header().Get("Content-Disposition")
	if !strings.HasPrefix(cd, "attachment; filename=") {
		t.Errorf("Content-Disposition = %q, want attachment; filename=...", cd)
	}
	if !strings.Contains(cd, "sec_events_") {
		t.Errorf("Content-Disposition = %q, want sec_events_ in filename", cd)
	}
	if !strings.HasSuffix(strings.Trim(cd, `"`), ".csv\"") && !strings.Contains(cd, ".csv") {
		t.Errorf("Content-Disposition = %q, want .csv extension", cd)
	}
}

func TestExportHandler_CSVHasHeaderAndRows(t *testing.T) {
	srv, cleanup := newTestServer(t, okF5Handler)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/export?window=1h", nil)
	rec := httptest.NewRecorder()
	srv.exportHandler(rec, req)

	rows, err := csv.NewReader(rec.Body).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV response: %v", err)
	}
	if len(rows) < 3 { // header + 2 events
		t.Fatalf("got %d CSV rows, want at least 3 (header + 2 events)", len(rows))
	}
	// Verify header columns
	wantCols := []string{"time", "src_ip", "country", "city", "vh_name", "app_type",
		"threat_level", "suspicion_score", "waf_sec_event_count", "req_count",
		"waf_suspicion_score", "summary_msg", "namespace", "tenant"}
	for i, col := range wantCols {
		if rows[0][i] != col {
			t.Errorf("CSV header[%d] = %q, want %q", i, rows[0][i], col)
		}
	}
	// Verify first data row src_ip
	if rows[1][1] != "1.2.3.4" {
		t.Errorf("CSV row 1 src_ip = %q, want 1.2.3.4", rows[1][1])
	}
}

func TestExportHandler_MethodNotAllowed(t *testing.T) {
	srv, cleanup := newTestServer(t, okF5Handler)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/export", nil)
	rec := httptest.NewRecorder()
	srv.exportHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}
