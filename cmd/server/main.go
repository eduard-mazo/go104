package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"go104/internal/api"
	"go104/internal/hub"
	"go104/internal/iec104"
	"go104/internal/metrics"
	"go104/internal/store"
	"go104/internal/ui"
)

// version is overridable at build time: -ldflags "-X main.version=1.4.0".
var version = "dev"

func main() {
	// Flags take precedence over env vars, which take precedence over defaults.
	// Empty default means "fall through to env/default" so the env path still works.
	portFlag := flag.String("port", "", "HTTP listen port (overrides $HTTP_PORT; default 8080)")
	dbFlag := flag.String("db", "", "SQLite database path (overrides $DB_PATH; default go104.db)")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	dbPath := pick(*dbFlag, env("DB_PATH", "go104.db"))
	httpAddr := ":" + pick(*portFlag, env("HTTP_PORT", "8080"))

	metrics.SetBuildInfo(version, vcsRevision(), runtime.Version())

	// Store
	st, err := store.New(dbPath)
	if err != nil {
		slog.Error("open store", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	// Hub (WebSocket broadcast)
	h := hub.New()
	go h.Run()

	// IEC104 master
	master := iec104.NewMaster(st, h)
	if err := master.StartAll(); err != nil {
		slog.Error("start lines", "err", err)
	}

	// HTTP router
	handlers := api.New(st, master, h)
	mux := http.NewServeMux()
	mux.Handle("/api/", handlers.Router())
	mux.Handle("/ws", handlers.Router())
	// Observability surface (root-mounted; exact patterns win over "/").
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok")) //nolint:errcheck
	})
	mux.HandleFunc("/readyz", readyHandler(st, master))
	mux.HandleFunc("/version", versionHandler())
	mux.Handle("/", spaHandler())

	srv := &http.Server{
		Addr:         httpAddr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server listening", "addr", httpAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down")
	master.StopAll()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown", "err", err)
	}
}

// spaHandler serves the embedded Vue dist, falling back to index.html
// for any path that doesn't map to a real file (SPA client-side routing).
func spaHandler() http.Handler {
	sub, err := fs.Sub(ui.FS, "dist")
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "frontend not built — run: make frontend", http.StatusServiceUnavailable)
		})
	}

	indexHTML, _ := ui.FS.ReadFile("dist/index.html")
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strip leading slash for fs.Open
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		f, err := sub.Open(path)
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Unknown path → serve index.html so Vue Router can handle it
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(indexHTML) //nolint:errcheck
	})
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// pick returns the first non-empty string (flag → env/default precedence).
func pick(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// readyHandler reports readiness: store reachable + line counts. 200 ready / 503 not.
func readyHandler(st store.Store, m *iec104.Master) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lines, err := st.ListLines()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]any{"ready": false, "store": err.Error()}) //nolint:errcheck
			return
		}
		active := 0
		for _, l := range lines {
			if m.IsRunning(l.ID) {
				active++
			}
		}
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"ready": true, "store": "ok",
			"lines_configured": len(lines), "lines_active": active,
		})
	}
}

func versionHandler() http.HandlerFunc {
	body := map[string]any{"version": version, "commit": vcsRevision(), "go": runtime.Version()}
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body) //nolint:errcheck
	}
}

// vcsRevision returns the embedded git commit (short) if the binary was built in a repo.
func vcsRevision() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			if s.Key == "vcs.revision" {
				if len(s.Value) > 12 {
					return s.Value[:12]
				}
				return s.Value
			}
		}
	}
	return "unknown"
}
