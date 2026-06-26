package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"go104/internal/hub"
	"go104/internal/iec104"
	"go104/internal/metrics"
	"go104/internal/store"
)

type Handlers struct {
	store  store.Store
	master *iec104.Master
	hub    *hub.Hub
}

func New(s store.Store, m *iec104.Master, h *hub.Hub) *Handlers {
	return &Handlers{store: s, master: m, hub: h}
}

func (h *Handlers) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
	}))
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(metricsMiddleware)

	r.Route("/api", func(r chi.Router) {
		// Lines
		r.Get("/lines", h.listLines)
		r.Post("/lines", h.createLine)
		r.Put("/lines/{id}", h.updateLine)
		r.Delete("/lines/{id}", h.deleteLine)
		r.Post("/lines/{id}/start", h.startLine)
		r.Post("/lines/{id}/stop", h.stopLine)
		r.Post("/lines/{id}/gi", h.triggerGI)

		// Signals
		r.Get("/signals", h.listSignals)
		r.Post("/signals", h.createSignal)
		r.Put("/signals/{id}", h.updateSignal)
		r.Delete("/signals/{id}", h.deleteSignal)

		// Live monitor values
		r.Get("/monitor", h.listDatapoints)

		// Commands
		r.Post("/commands", h.sendCommand)

		// All signals (for SCADA signal picker)
		r.Get("/signals/all", h.listAllSignals)

		// Signal history (time-series for chart modal)
		r.Get("/signals/{id}/history", h.getSignalHistory)

		// SCADA views
		r.Get("/scada/views", h.listScadaViews)
		r.Post("/scada/views", h.createScadaView)
		r.Get("/scada/views/{id}", h.getScadaView)
		r.Put("/scada/views/{id}", h.updateScadaView)
		r.Delete("/scada/views/{id}", h.deleteScadaView)
	})

	// WebSocket
	r.Get("/ws", h.wsHandler)

	return r
}

// metricsMiddleware records request count + duration per matched route (chi
// pattern, not raw path → bounded cardinality).
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(ww, r)
		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "other"
		}
		metrics.HTTPRequest(route, r.Method, ww.Status(), time.Since(start).Seconds())
	})
}
