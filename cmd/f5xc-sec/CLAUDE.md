# Package: cmd/f5xc-sec

## Purpose
CLI entry point for the F5 XC Security Events tool.

## Key File
- main.go — parses CLI flags, loads config, dispatches to CLI or web server mode

## CLI Flags
- `--window`    int     Time window in hours, 1–24 (default 1)
- `--namespace` string  F5 XC namespace (overrides F5XC_NAMESPACE env var)
- `--lb`        string  HTTP Load Balancer name to filter events
- `--serve`     bool    Start web server mode instead of printing JSON
- `--port`      int     Web server port (default 8080)
- `--export`    bool    Export events as CSV to stdout
- `--timeout`   int     HTTP client timeout in seconds (default 30)

## Implementation Status: MOSTLY COMPLETE — --export stub remaining (Prompt 4)

### Wired as of Prompt 3
- `--serve`: creates `web.NewServer(client, cfg)` and calls `Start(*port)` — blocks
- `--timeout`: applied via `client.WithTimeout(time.Duration(*timeout) * time.Second)`
- `--namespace`: overrides `cfg.Namespace` after `config.Load()`
- CLI default mode: `client.FetchEvents(context.Background(), ...)` → `json.NewEncoder(os.Stdout).Encode(events)`

### --export wired (Prompt 4) — COMPLETE
- Flag variable is `doExport *bool` (not `export`, which would shadow the import)
- Both `--export` and default JSON mode share a single `FetchEvents` call
- Flow: serve → (fetch) → export CSV → (default) JSON print
