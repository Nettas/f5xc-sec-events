package web

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nettas12/f5xc-sec-events/internal/api"
	"github.com/nettas12/f5xc-sec-events/internal/config"
)

// ── Test helpers ──────────────────────────────────────────────────────────────

// mockF5Response is a minimal valid F5 XC API response with two events.
var mockF5Response = `{
  "events": [
    {
      "time": "2026-04-16T10:00:00Z",
      "src_ip": "1.2.3.4",
      "req_path": "/admin",
      "method": "GET",
      "response_code": 403,
      "req_id": "req-001",
      "waf_action": "BLOCK",
      "attack_type": "SQL_INJECTION",
      "severity": "HIGH",
      "virtual_host": "my-lb"
    },
    {
      "time": "2026-04-16T10:05:00Z",
      "src_ip": "5.6.7.8",
      "req_path": "/login",
      "method": "POST",
      "response_code": 200,
      "req_id": "req-002",
      "waf_action": "ALLOW",
      "attack_type": "",
      "severity": "LOW",
      "virtual_host": "my-lb"
    }
  ]
}`

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
	if events[1].WAFAction != "ALLOW" {
		t.Errorf("events[1].WAFAction = %q, want ALLOW", events[1].WAFAction)
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
	wantCols := []string{"time", "src_ip", "method", "req_path", "response_code",
		"waf_action", "attack_type", "severity", "virtual_host", "req_id"}
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
