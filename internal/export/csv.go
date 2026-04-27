package export

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/Nettas/f5xc-sec-events/internal/api"
)

// WriteCSV serialises events to w in CSV format.
func WriteCSV(w io.Writer, events []api.SecurityEvent) error {
	cw := csv.NewWriter(w)

	header := []string{
		"time", "src_ip", "country", "city", "vh_name", "app_type",
		"threat_level", "suspicion_score", "waf_sec_event_count", "req_count",
		"waf_suspicion_score", "summary_msg", "namespace", "tenant",
	}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	for _, e := range events {
		row := []string{
			e.Time,
			e.SrcIP,
			e.Country,
			e.City,
			e.VhName,
			e.AppType,
			e.ThreatLevel,
			fmt.Sprintf("%g", e.SuspicionScore),
			fmt.Sprintf("%d", e.WafSecEventCount),
			fmt.Sprintf("%d", e.ReqCount),
			fmt.Sprintf("%g", e.WafSuspicionScore),
			e.SummaryMsg,
			e.Namespace,
			e.Tenant,
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("write CSV row: %w", err)
		}
	}

	cw.Flush()
	return cw.Error()
}
