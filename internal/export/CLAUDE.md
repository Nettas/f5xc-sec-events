# Package: internal/export

## Purpose
Serialize a []api.SecurityEvent slice to CSV format.

## CSV Columns (14 columns, confirmed against live API fields 2026-04-20)
    time, src_ip, country, city, vh_name, app_type,
    threat_level, suspicion_score, waf_sec_event_count, req_count,
    waf_suspicion_score, summary_msg, namespace, tenant

Note: old columns (method, req_path, response_code, waf_action, attack_type,
severity, virtual_host, req_id) were removed — they do not exist in the live API response.

## Requirements
- Use encoding/csv (stdlib only, no third-party libs)
- WriteCSV(w io.Writer, events []api.SecurityEvent) error
- Must flush the csv.Writer before returning
- Float fields use fmt.Sprintf("%g", ...) — omits trailing zeros

## Implementation Status: COMPLETE

### csv.go — DONE
- `WriteCSV(w io.Writer, events []api.SecurityEvent) error`
- Writes header row then one row per event; calls `cw.Flush()` and returns `cw.Error()`
- Numeric fields: `%g` for floats, `%d` for ints
- Integrated into `web/handlers.go` exportHandler

### csv_test.go — DONE
- `TestWriteCSV_HeaderRow` — asserts all 14 column names in order
- `TestWriteCSV_RowCount` — header + 3 data rows = 4 total
- `TestWriteCSV_DataRows` — asserts all 14 field values per row
- `TestWriteCSV_EmptySlice` — empty input produces header-only output
