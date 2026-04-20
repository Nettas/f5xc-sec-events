# Package: internal/export

## Purpose
Serialize a []api.SecurityEvent slice to CSV format.

## Requirements
- Use encoding/csv (stdlib only, no third-party libs)
- CSV columns (in order):
    time, src_ip, method, req_path, response_code, waf_action,
    attack_type, severity, virtual_host, req_id
- WriteCSV(w io.Writer, events []api.SecurityEvent) error
- Must flush the csv.Writer before returning
- Dates in CSV should be human-readable (RFC3339 format)

## Implementation Status: PARTIALLY COMPLETE (csv.go done; CLI mode + tests pending Prompt 4)

### csv.go — DONE
- `WriteCSV(w io.Writer, events []api.SecurityEvent) error`
- Writes header row then one row per event; calls `cw.Flush()` and returns `cw.Error()`
- `response_code` converted with `fmt.Sprintf("%d", e.ResponseCode)`
- Already integrated into `web/handlers.go` exportHandler

### csv_test.go — DONE (Prompt 4)
- `TestWriteCSV_HeaderRow` — asserts all 10 column names in order
- `TestWriteCSV_RowCount` — header + 3 data rows = 4 total
- `TestWriteCSV_DataRows` — asserts all 10 field values per row
- `TestWriteCSV_EmptySlice` — empty input produces header-only output
