package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go104/internal/models"
)

func (h *Handlers) listSignals(w http.ResponseWriter, r *http.Request) {
	lineIDStr := r.URL.Query().Get("line_id")
	if lineIDStr == "" {
		jsonError(w, "line_id required", http.StatusBadRequest)
		return
	}
	lineID, err := strconv.ParseInt(lineIDStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid line_id", http.StatusBadRequest)
		return
	}
	sigs, err := h.store.ListSignals(lineID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if sigs == nil {
		sigs = []models.Signal{}
	}
	jsonOK(w, sigs)
}

func (h *Handlers) createSignal(w http.ResponseWriter, r *http.Request) {
	var sig models.Signal
	if err := json.NewDecoder(r.Body).Decode(&sig); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if sig.Name == "" || sig.IOA == 0 || sig.LineID == 0 {
		jsonError(w, "name, ioa and line_id are required", http.StatusBadRequest)
		return
	}
	if sig.Scale == 0 {
		sig.Scale = 1.0
	}
	if err := h.store.CreateSignal(&sig); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.master.RefreshSignals(sig.LineID)
	jsonOK(w, sig)
}

func (h *Handlers) updateSignal(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var sig models.Signal
	if err := json.NewDecoder(r.Body).Decode(&sig); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	sig.ID = id
	if sig.Scale == 0 {
		sig.Scale = 1.0
	}
	if err := h.store.UpdateSignal(&sig); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.master.RefreshSignals(sig.LineID)
	jsonOK(w, sig)
}

func (h *Handlers) deleteSignal(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	sig, err := h.store.GetSignal(id)
	if err != nil || sig == nil {
		jsonError(w, "signal not found", http.StatusNotFound)
		return
	}
	if err := h.store.DeleteSignal(id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.master.RefreshSignals(sig.LineID)
	jsonOK(w, map[string]any{"deleted": id})
}

func (h *Handlers) listDatapoints(w http.ResponseWriter, r *http.Request) {
	lineIDStr := r.URL.Query().Get("line_id")
	if lineIDStr == "" {
		jsonError(w, "line_id required", http.StatusBadRequest)
		return
	}
	lineID, err := strconv.ParseInt(lineIDStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid line_id", http.StatusBadRequest)
		return
	}
	dps, err := h.store.ListDatapoints(lineID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, dps)
}
