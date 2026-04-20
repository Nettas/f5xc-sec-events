package export

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"

	"github.com/nettas12/f5xc-sec-events/internal/api"
)

var mockEvents = []api.SecurityEvent{
	{
		Time:         "2026-04-16T10:00:00Z",
		SrcIP:        "1.2.3.4",
		Method:       "GET",
		ReqPath:      "/admin",
		ResponseCode: 403,
		WAFAction:    "BLOCK",
		AttackType:   "SQL_INJECTION",
		Severity:     "HIGH",
		VirtualHost:  "my-lb",
		ReqID:        "req-001",
	},
	{
		Time:         "2026-04-16T10:05:00Z",
		SrcIP:        "5.6.7.8",
		Method:       "POST",
		ReqPath:      "/login",
		ResponseCode: 200,
		WAFAction:    "ALLOW",
		AttackType:   "",
		Severity:     "LOW",
		VirtualHost:  "my-lb",
		ReqID:        "req-002",
	},
	{
		Time:         "2026-04-16T10:10:00Z",
		SrcIP:        "9.10.11.12",
		Method:       "GET",
		ReqPath:      "/etc/passwd",
		ResponseCode: 403,
		WAFAction:    "BLOCK",
		AttackType:   "PATH_TRAVERSAL",
		Severity:     "CRITICAL",
		VirtualHost:  "my-lb",
		ReqID:        "req-003",
	},
}

// TestWriteCSV_HeaderRow verifies the header is the first row and contains all 10 columns.
func TestWriteCSV_HeaderRow(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCSV(&buf, mockEvents); err != nil {
		t.Fatalf("WriteCSV error: %v", err)
	}

	r := csv.NewReader(&buf)
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}

	wantHeader := []string{
		"time", "src_ip", "method", "req_path", "response_code",
		"waf_action", "attack_type", "severity", "virtual_host", "req_id",
	}
	if len(rows) == 0 {
		t.Fatal("CSV output is empty")
	}
	header := rows[0]
	if len(header) != len(wantHeader) {
		t.Fatalf("header has %d columns, want %d", len(header), len(wantHeader))
	}
	for i, col := range wantHeader {
		if header[i] != col {
			t.Errorf("header[%d] = %q, want %q", i, header[i], col)
		}
	}
}

// TestWriteCSV_RowCount verifies there is one data row per event plus the header.
func TestWriteCSV_RowCount(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCSV(&buf, mockEvents); err != nil {
		t.Fatalf("WriteCSV error: %v", err)
	}

	r := csv.NewReader(&buf)
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}

	want := len(mockEvents) + 1 // +1 for header
	if len(rows) != want {
		t.Errorf("got %d rows (incl. header), want %d", len(rows), want)
	}
}

// TestWriteCSV_DataRows verifies all 10 column values for each mock event.
func TestWriteCSV_DataRows(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCSV(&buf, mockEvents); err != nil {
		t.Fatalf("WriteCSV error: %v", err)
	}

	r := csv.NewReader(&buf)
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}

	// rows[0] is header; data starts at rows[1]
	type wantRow struct {
		time, srcIP, method, reqPath, code, action, atkType, severity, vhost, reqID string
	}
	want := []wantRow{
		{"2026-04-16T10:00:00Z", "1.2.3.4",    "GET",  "/admin",       "403", "BLOCK", "SQL_INJECTION",  "HIGH",     "my-lb", "req-001"},
		{"2026-04-16T10:05:00Z", "5.6.7.8",    "POST", "/login",        "200", "ALLOW", "",               "LOW",      "my-lb", "req-002"},
		{"2026-04-16T10:10:00Z", "9.10.11.12", "GET",  "/etc/passwd",   "403", "BLOCK", "PATH_TRAVERSAL", "CRITICAL", "my-lb", "req-003"},
	}

	for i, w := range want {
		row := rows[i+1]
		checks := []struct{ got, want, col string }{
			{row[0], w.time,     "time"},
			{row[1], w.srcIP,    "src_ip"},
			{row[2], w.method,   "method"},
			{row[3], w.reqPath,  "req_path"},
			{row[4], w.code,     "response_code"},
			{row[5], w.action,   "waf_action"},
			{row[6], w.atkType,  "attack_type"},
			{row[7], w.severity, "severity"},
			{row[8], w.vhost,    "virtual_host"},
			{row[9], w.reqID,    "req_id"},
		}
		for _, c := range checks {
			if c.got != c.want {
				t.Errorf("row %d col %s: got %q, want %q", i+1, c.col, c.got, c.want)
			}
		}
	}
}

// TestWriteCSV_EmptySlice verifies that an empty event slice produces only a header row.
func TestWriteCSV_EmptySlice(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCSV(&buf, []api.SecurityEvent{}); err != nil {
		t.Fatalf("WriteCSV error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("empty input: got %d lines, want 1 (header only)", len(lines))
	}
}
