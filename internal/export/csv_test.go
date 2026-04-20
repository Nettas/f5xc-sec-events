package export

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"

	"github.com/Nettas/f5xc-sec-events/internal/api"
)

var mockEvents = []api.SecurityEvent{
	{
		Time:             "2026-04-16T10:00:00Z",
		SrcIP:            "1.2.3.4",
		Country:          "US",
		City:             "New York",
		VhName:           "ves-io-s-iannetta-my-lb",
		AppType:          "web",
		ThreatLevel:      "high",
		SuspicionScore:   85.5,
		WafSecEventCount: 3,
		ReqCount:         10,
		WafSuspicionScore: 75.0,
		SummaryMsg:       "SQL injection attempt",
		Namespace:        "s-iannetta",
		Tenant:           "f5-sa",
	},
	{
		Time:             "2026-04-16T10:05:00Z",
		SrcIP:            "5.6.7.8",
		Country:          "GB",
		City:             "London",
		VhName:           "ves-io-s-iannetta-my-lb",
		AppType:          "web",
		ThreatLevel:      "low",
		SuspicionScore:   10.0,
		WafSecEventCount: 0,
		ReqCount:         5,
		WafSuspicionScore: 0,
		SummaryMsg:       "",
		Namespace:        "s-iannetta",
		Tenant:           "f5-sa",
	},
	{
		Time:             "2026-04-16T10:10:00Z",
		SrcIP:            "9.10.11.12",
		Country:          "CN",
		City:             "Beijing",
		VhName:           "ves-io-s-iannetta-my-lb",
		AppType:          "web",
		ThreatLevel:      "critical",
		SuspicionScore:   99.0,
		WafSecEventCount: 7,
		ReqCount:         20,
		WafSuspicionScore: 95.0,
		SummaryMsg:       "Path traversal attempt",
		Namespace:        "s-iannetta",
		Tenant:           "f5-sa",
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
		"time", "src_ip", "country", "city", "vh_name", "app_type",
		"threat_level", "suspicion_score", "waf_sec_event_count", "req_count",
		"waf_suspicion_score", "summary_msg", "namespace", "tenant",
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
		time, srcIP, country, city, vhName, appType, threatLevel,
		suspicionScore, wafSecEventCount, reqCount, wafSuspicionScore,
		summaryMsg, namespace, tenant string
	}
	want := []wantRow{
		{"2026-04-16T10:00:00Z", "1.2.3.4",    "US", "New York", "ves-io-s-iannetta-my-lb", "web", "high",     "85.5", "3", "10", "75",  "SQL injection attempt",    "s-iannetta", "f5-sa"},
		{"2026-04-16T10:05:00Z", "5.6.7.8",    "GB", "London",   "ves-io-s-iannetta-my-lb", "web", "low",      "10",   "0", "5",  "0",   "",                         "s-iannetta", "f5-sa"},
		{"2026-04-16T10:10:00Z", "9.10.11.12", "CN", "Beijing",  "ves-io-s-iannetta-my-lb", "web", "critical", "99",   "7", "20", "95",  "Path traversal attempt",   "s-iannetta", "f5-sa"},
	}

	for i, w := range want {
		row := rows[i+1]
		checks := []struct{ got, want, col string }{
			{row[0],  w.time,             "time"},
			{row[1],  w.srcIP,            "src_ip"},
			{row[2],  w.country,          "country"},
			{row[3],  w.city,             "city"},
			{row[4],  w.vhName,           "vh_name"},
			{row[5],  w.appType,          "app_type"},
			{row[6],  w.threatLevel,      "threat_level"},
			{row[7],  w.suspicionScore,   "suspicion_score"},
			{row[8],  w.wafSecEventCount, "waf_sec_event_count"},
			{row[9],  w.reqCount,         "req_count"},
			{row[10], w.wafSuspicionScore,"waf_suspicion_score"},
			{row[11], w.summaryMsg,       "summary_msg"},
			{row[12], w.namespace,        "namespace"},
			{row[13], w.tenant,           "tenant"},
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
