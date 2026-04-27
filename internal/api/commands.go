package api

import (
	"encoding/json"
	"net/http"

	"go104/internal/models"
)

func (h *Handlers) sendCommand(w http.ResponseWriter, r *http.Request) {
	var cmd models.Command
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if cmd.LineID == 0 || cmd.IOA == 0 || cmd.TypeID == 0 {
		jsonError(w, "line_id, ioa and type_id are required", http.StatusBadRequest)
		return
	}
	if err := h.master.SendCommand(cmd); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, map[string]any{"queued": true})
}
