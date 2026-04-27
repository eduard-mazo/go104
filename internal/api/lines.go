package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go104/internal/models"
)

func (h *Handlers) listLines(w http.ResponseWriter, r *http.Request) {
	lines, err := h.store.ListLines()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if lines == nil {
		lines = []models.Line{}
	}
	// Enrich with runtime state
	for i := range lines {
		lines[i].State = h.master.LineState(lines[i].ID)
		lines[i].RxCount, lines[i].TxCount = h.master.LineStats(lines[i].ID)
		lines[i].Addr = lines[i].Host + ":" + strconv.Itoa(lines[i].Port)
	}
	jsonOK(w, lines)
}

func (h *Handlers) createLine(w http.ResponseWriter, r *http.Request) {
	var l models.Line
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if l.Name == "" || l.Host == "" || l.Port == 0 {
		jsonError(w, "name, host and port are required", http.StatusBadRequest)
		return
	}
	// Defaults
	if l.K == 0 {
		l.K = 12
	}
	if l.W == 0 {
		l.W = 8
	}
	if l.T1MS == 0 {
		l.T1MS = 15000
	}
	if l.T2MS == 0 {
		l.T2MS = 10000
	}
	if l.T3MS == 0 {
		l.T3MS = 20000
	}
	if l.CommonAddr == 0 {
		l.CommonAddr = 1
	}
	if err := h.store.CreateLine(&l); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if l.Enabled {
		_ = h.master.StartLine(l)
	}
	jsonOK(w, l)
}

func (h *Handlers) updateLine(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var l models.Line
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	l.ID = id
	if err := h.store.UpdateLine(&l); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Restart worker with new config if it was running
	if h.master.IsRunning(id) || l.Enabled {
		_ = h.master.RestartLine(l)
	}
	jsonOK(w, l)
}

func (h *Handlers) deleteLine(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	_ = h.master.StopLine(id)
	if err := h.store.DeleteLine(id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]any{"deleted": id})
}

func (h *Handlers) startLine(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	line, err := h.store.GetLine(id)
	if err != nil || line == nil {
		jsonError(w, "line not found", http.StatusNotFound)
		return
	}
	if h.master.IsRunning(id) {
		jsonError(w, "already running", http.StatusConflict)
		return
	}
	if err := h.master.StartLine(*line); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]any{"started": id})
}

func (h *Handlers) stopLine(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.master.StopLine(id); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, map[string]any{"stopped": id})
}

func (h *Handlers) triggerGI(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.master.SendGI(id); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, map[string]any{"gi_triggered": id})
}

// --- helpers ---

func pathInt64(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, key), 10, 64)
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}
