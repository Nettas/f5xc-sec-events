package web

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/nettas12/f5xc-sec-events/internal/api"
	"github.com/nettas12/f5xc-sec-events/internal/config"
)

//go:embed static
var staticFiles embed.FS

// Server is an HTTP server that serves the dashboard UI and exposes the events API.
type Server struct {
	client *api.Client
	cfg    *config.Config
	mux    *http.ServeMux
}

// NewServer creates a Server wired to the given API client and config.
func NewServer(client *api.Client, cfg *config.Config) *Server {
	s := &Server{
		client: client,
		cfg:    cfg,
		mux:    http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/config", s.configHandler)
	s.mux.HandleFunc("/api/events", s.eventsHandler)
	s.mux.HandleFunc("/api/export", s.exportHandler)

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		// embed.FS.Sub only errors if the path doesn't exist — panic is appropriate.
		panic(fmt.Sprintf("web: embed static subtree: %v", err))
	}
	s.mux.Handle("/", http.FileServer(http.FS(staticFS)))
}

// Start listens on the given port and blocks until the server exits.
func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	fmt.Fprintf(os.Stderr, "Dashboard ready → http://localhost%s\n", addr)
	return http.ListenAndServe(addr, s.mux)
}
