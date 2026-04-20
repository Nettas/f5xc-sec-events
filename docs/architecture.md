# Architecture

## Overview

```
┌─────────────────────────────────────────────┐
│  cmd/f5xc-sec/main.go  (CLI entry point)    │
│  - flag parsing                             │
│  - config loading                           │
│  - dispatch: CLI | serve | export           │
└────────────┬──────────────┬─────────────────┘
             │              │
     ┌───────▼──────┐  ┌────▼──────────────┐
     │ internal/api │  │ web/               │
     │ client.go    │  │ server.go          │
     │ events.go    │  │ handlers.go        │
     │ models.go    │  │ static/ (embedded) │
     └───────┬──────┘  └────────────────────┘
             │
     ┌───────▼──────────┐
     │ internal/export  │
     │ csv.go           │
     └──────────────────┘
```

## Data Flow

1. CLI/web receives `--window` (1h | 24h) and `--lb` (load balancer name)
2. `internal/config` loads env vars (API key, tenant, namespace)
3. `internal/api.FetchEvents` POSTs to F5 XC API with computed time range
4. Response deserialized into `[]SecurityEvent`
5. Output routed to: JSON stdout | web JSON response | CSV (stdout or HTTP)
