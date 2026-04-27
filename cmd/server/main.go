package main

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go104/internal/api"
	"go104/internal/hub"
	"go104/internal/iec104"
	"go104/internal/store"
	"go104/internal/ui"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	dbPath := env("DB_PATH", "go104.db")
	httpAddr := ":" + env("HTTP_PORT", "8080")

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
