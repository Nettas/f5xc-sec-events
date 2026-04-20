package export

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/nettas12/f5xc-sec-events/internal/api"
)

// WriteCSV serialises events to w in CSV format.
// Columns: time, src_ip, method, req_path, response_code, waf_action,
//
//	attack_type, severity, virtual_host, req_id
func WriteCSV(w io.Writer, events []api.SecurityEvent) error {
	cw := csv.NewWriter(w)

	header := []string{
		"time", "src_ip", "method", "req_path", "response_code",
		"waf_action", "attack_type", "severity", "virtual_host", "req_id",
	}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	for _, e := range events {
		row := []string{
			e.Time,
			e.SrcIP,
			e.Method,
			e.ReqPath,
			fmt.Sprintf("%d", e.ResponseCode),
			e.WAFAction,
			e.AttackType,
			e.Severity,
			e.VirtualHost,
			e.ReqID,
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("write CSV row: %w", err)
		}
	}

	cw.Flush()
	return cw.Error()
}
