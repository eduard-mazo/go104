package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go104/internal/models"
)

func (h *Handlers) listScadaViews(w http.ResponseWriter, r *http.Request) {
	views, err := h.store.ListScadaViews()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if views == nil {
		views = []models.ScadaView{}
	}
	jsonOK(w, views)
}

func (h *Handlers) getScadaView(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	v, err := h.store.GetScadaView(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if v == nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	jsonOK(w, v)
}

func (h *Handlers) createScadaView(w http.ResponseWriter, r *http.Request) {
	var v models.ScadaView
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if v.Name == "" {
		jsonError(w, "name required", http.StatusBadRequest)
		return
	}
	if v.Width == 0 {
		v.Width = 1400
	}
	if v.Height == 0 {
		v.Height = 900
	}
	if len(v.Elements) == 0 {
		v.Elements = json.RawMessage("[]")
	}
	if len(v.Lines) == 0 {
		v.Lines = json.RawMessage("[]")
	}
	if err := h.store.CreateScadaView(&v); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, v)
}

func (h *Handlers) updateScadaView(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var v models.ScadaView
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	v.ID = id
	if len(v.Elements) == 0 {
		v.Elements = json.RawMessage("[]")
	}
	if len(v.Lines) == 0 {
		v.Lines = json.RawMessage("[]")
	}
	if err := h.store.UpdateScadaView(&v); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, v)
}

func (h *Handlers) deleteScadaView(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.store.DeleteScadaView(id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]any{"deleted": id})
}

func (h *Handlers) listAllSignals(w http.ResponseWriter, r *http.Request) {
	sigs, err := h.store.ListAllSignals()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if sigs == nil {
		sigs = []models.Signal{}
	}
	jsonOK(w, sigs)
}

func (h *Handlers) getSignalHistory(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()
	now := float64(time.Now().UnixNano()) / 1e9
	from := now - 3600 // default: last 1h
	to := now
	limit := 2000

	if v := q.Get("from"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			from = f
		}
	}
	if v := q.Get("to"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			to = f
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	pts, err := h.store.QueryHistory(id, from, to, limit)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pts == nil {
		pts = []models.HistoryPoint{}
	}
	jsonOK(w, pts)
}
