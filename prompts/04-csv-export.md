Read internal/export/CLAUDE.md first. Then:

1. Implement internal/export/csv.go per spec (columns, order, flush)
2. Integrate into web/handlers.go exportHandler
3. Also support CLI export mode:
   - When `--export` flag is set, run FetchEvents and pipe through WriteCSV to stdout
   - Example: `go run ./cmd/f5xc-sec --window 24h --lb my-lb --export > events.csv`
4. Write a unit test for WriteCSV with 3 mock SecurityEvent records,
   verify CSV header row and data rows
