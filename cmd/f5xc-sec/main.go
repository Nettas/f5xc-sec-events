package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nettas12/f5xc-sec-events/internal/api"
	"github.com/nettas12/f5xc-sec-events/internal/config"
	"github.com/nettas12/f5xc-sec-events/internal/export"
	"github.com/nettas12/f5xc-sec-events/web"
)

func main() {
	window    := flag.String("window",    "1h",  "Time window: 1h or 24h")
	namespace := flag.String("namespace", "",    "F5 XC namespace (overrides F5XC_NAMESPACE)")
	lb        := flag.String("lb",        "",    "HTTP Load Balancer name to filter events")
	serve     := flag.Bool("serve",       false, "Start web server mode")
	port      := flag.Int("port",         4000,  "Web server port (used with --serve)")
	doExport  := flag.Bool("export",      false, "Export events as CSV to stdout")
	timeout   := flag.Int("timeout",      30,    "HTTP client timeout in seconds")
	flag.Parse()

	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	cfg, err := config.Load()
	if err != nil {
		if !*serve {
			// CLI and export modes require the API key up-front.
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		// Web server mode: start without env key — user can paste it in the dashboard.
		fmt.Fprintf(os.Stderr, "warning: %v — enter API key in the dashboard UI\n", err)
		cfg = config.Defaults()
	}
	if *namespace != "" {
		cfg.Namespace = *namespace
	}

	client := api.NewClient(cfg.Tenant, cfg.APIKey).
		WithTimeout(time.Duration(*timeout) * time.Second)

	// ── Web server mode ────────────────────────────────────────────
	if *serve {
		srv := web.NewServer(client, cfg)
		if err := srv.Start(*port); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// ── Fetch events (shared by CLI and export modes) ─────────────
	events, err := client.FetchEvents(context.Background(), cfg.Namespace, *lb, *window)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// ── CSV export mode ────────────────────────────────────────────
	if *doExport {
		if err := export.WriteCSV(os.Stdout, events); err != nil {
			fmt.Fprintf(os.Stderr, "export error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// ── Default CLI mode: print JSON ───────────────────────────────
	if err := json.NewEncoder(os.Stdout).Encode(events); err != nil {
		fmt.Fprintf(os.Stderr, "encode: %v\n", err)
		os.Exit(1)
	}
}
