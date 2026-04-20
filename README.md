# F5 XC Security Events Tool

A Go CLI + browser dashboard that pulls WAF/security events from F5 Distributed Cloud,
visualizes them in real-time, and exports them to CSV.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Getting an F5 XC API Key](#getting-an-f5-xc-api-key)
- [Quick Start — Running Locally](#quick-start--running-locally)
- [Running — Web Dashboard (recommended)](#running--web-dashboard-recommended)
- [Running — CLI Mode](#running--cli-mode-prints-json)
- [Running — CSV Export](#running--csv-export-to-file)
- [All CLI Flags](#all-cli-flags)
- [Environment Variables](#environment-variables)
- [Development](#development)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

| Requirement | Version | Notes |
|---|---|---|
| Go | 1.22+ | https://go.dev/dl/ — free download |
| Git | any | To clone the repo |
| F5 XC API key | — | See [Getting an API Key](#getting-an-f5-xc-api-key) |

**Verify your Go installation:**
```bash
go version
# Should print: go version go1.22.x ...
```

> **macOS / Linux note:** If `go` is not found, add it to your PATH:
> ```bash
> export PATH=$PATH:/usr/local/go/bin    # typical macOS/Linux install
> # or wherever you installed Go
> ```
> On this dev machine Go lives at `/home/coder/go/bin/go` — use the full path if needed.

---

## Getting an F5 XC API Key

1. Log in to the F5 Distributed Cloud Console: https://f5-sa.console.ves.volterra.io
2. Click your **user avatar** (top-right) → **Account Settings**
3. Select **Credentials** in the left sidebar
4. Click **Add Credentials**
   - Name: `sec-events-tool` (or any name you like)
   - Credential Type: **API Token**
   - Expiry: set as needed
5. Click **Generate** — copy the token immediately (it is shown only once)

---

## Quick Start — Running Locally

Follow these steps from scratch on any machine with Go 1.22+ installed.

### Step 1 — Clone the repository

```bash
git clone https://github.com/nettas12/f5xc-sec-events.git
cd f5xc-sec-events
```

### Step 2 — Configure your API key

```bash
# Copy the example env file
cp .env.example .env

# Edit .env and fill in your API key
#   F5XC_API_KEY=<paste-your-token-here>
#   F5XC_TENANT=f5-sa              ← leave as-is unless your tenant differs
#   F5XC_NAMESPACE=s-iannetta      ← change to your namespace if needed
```

You can use any text editor (`nano .env`, `vim .env`, or open it in VS Code).

### Step 3 — Build the binary

```bash
go build -o bin/f5xc-sec ./cmd/f5xc-sec
```

This produces a single self-contained binary at `bin/f5xc-sec` (no runtime dependencies,
all web assets are embedded).

Verify:
```bash
./bin/f5xc-sec --help
```

### Step 4 — Run the web dashboard

```bash
# Load your API key from .env, then start the dashboard
source .env
./bin/f5xc-sec --serve --port 4000
```

Open **http://localhost:4000** in your browser. You should see the dark-theme dashboard.

**First-time setup in the browser:**
1. Expand the **Connection** section in the left sidebar
2. Paste your API token into the **API Key** field
3. Confirm the **Namespace** shows your namespace (e.g. `s-iannetta`)
4. Enter your **Load Balancer** name (e.g. `my-lb`) — leave blank for all LBs
5. Click **⟳ Refresh**

> The API key is stored in `sessionStorage` only — it is never saved to disk and clears
> automatically when you close the browser tab.

---

## Running — Web Dashboard (recommended)

The dashboard lets you enter your API key, namespace, and LB name directly in the
browser — no environment variables required at startup.

```bash
# Simplest: no env vars needed — enter key in the UI
./bin/f5xc-sec --serve --port 4000

# Pre-load env vars so the server already has the key (skips manual entry in the UI)
source .env && ./bin/f5xc-sec --serve --port 4000

# Custom port
./bin/f5xc-sec --serve --port 8080
```

Open **http://localhost:4000** (or whichever port you chose).

**Dashboard controls:**
- **1h / 24h** toggle — switch the time window and re-fetch automatically
- **↓ Export CSV** — download a timestamped CSV of the currently displayed events
- Click any **column header** to sort the event table (click again to reverse)
- **⟳ Refresh** — manually re-fetch events

---

## Running — CLI Mode (prints JSON)

Useful for scripting, automation, or piping into `jq`.

```bash
source .env

# Last 1 hour, all load balancers
./bin/f5xc-sec --window 1h

# Last 24 hours, specific LB
./bin/f5xc-sec --window 24h --lb my-lb

# Pretty-print with jq
./bin/f5xc-sec --window 1h --lb my-lb | jq '.[].src_ip'

# Override namespace on the command line
./bin/f5xc-sec --window 1h --namespace s-iannetta --lb my-lb

# Increase timeout for slow connections
./bin/f5xc-sec --window 24h --timeout 60
```

Or run without building first using `go run`:
```bash
source .env
go run ./cmd/f5xc-sec --window 1h --lb my-lb
```

---

## Running — CSV Export to File

Export events directly to a CSV file for Excel or further analysis.

```bash
source .env

# Last 1 hour → events.csv
./bin/f5xc-sec --window 1h --lb my-lb --export > events.csv

# Last 24 hours → timestamped filename
./bin/f5xc-sec --window 24h --lb my-lb --export > "events_$(date +%Y%m%d_%H%M%S).csv"

# Pipe through head to preview
./bin/f5xc-sec --window 1h --export | head -5
```

**CSV columns:** `time, src_ip, method, req_path, response_code, waf_action, attack_type, severity, virtual_host, req_id`

---

## All CLI Flags

| Flag | Default | Description |
|---|---|---|
| `--window` | `1h` | Time window: `1h` or `24h` |
| `--namespace` | `$F5XC_NAMESPACE` | F5 XC namespace (overrides env var) |
| `--lb` | _(none)_ | HTTP Load Balancer name to filter events |
| `--serve` | `false` | Start web dashboard mode |
| `--port` | `4000` | Web server port (used with `--serve`) |
| `--export` | `false` | Write events as CSV to stdout |
| `--timeout` | `30` | HTTP client timeout in seconds |

---

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `F5XC_API_KEY` | Yes (CLI/export) | — | F5 XC API token |
| `F5XC_TENANT` | No | `f5-sa` | F5 XC tenant name |
| `F5XC_NAMESPACE` | No | `s-iannetta` | F5 XC namespace |

In **web server mode** (`--serve`), `F5XC_API_KEY` is optional — you can enter the key
in the browser UI instead. In **CLI and export mode** it is required.

---

## Development

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Lint / vet
go vet ./...

# Build and run in one step (no binary output)
source .env
go run ./cmd/f5xc-sec --serve --port 4000
```

**Test coverage:** 20 tests across `internal/api`, `internal/export`, and `web` packages.

---

## Troubleshooting

**`go: command not found`**
→ Go is not on your PATH. Either add it: `export PATH=$PATH:/usr/local/go/bin`
→ Or use the full path: `/usr/local/go/bin/go build ./...`

**`config error: F5XC_API_KEY environment variable is required`**
→ In web mode this is just a warning — paste the key in the dashboard UI.
→ In CLI/export mode: `export F5XC_API_KEY=your-token` or `source .env`

**`fetch events: F5 XC API returned HTTP 401`**
→ API key is invalid or expired. Generate a new one in the F5 XC console.

**`fetch events: F5 XC API returned HTTP 403`**
→ API key lacks permission for the namespace. Check your F5 XC role assignments.

**Dashboard shows "No events found"**
→ Try widening to `24h`. Confirm the LB name exactly matches the name in F5 XC.
→ Leave the LB field blank to return all events across all load balancers.

**`dial tcp: connection refused`**
→ Check network connectivity to `f5-sa.console.ves.volterra.io`.
→ Try increasing the timeout: `--timeout 60`

**Port already in use**
→ Change the port: `./bin/f5xc-sec --serve --port 8888`
→ Or find what's using it: `lsof -i :4000`
