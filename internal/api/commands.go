package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go104/internal/metrics"
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
	err := h.master.SendCommand(cmd)
	metrics.Command(h.lineLabel(cmd.LineID), cmd.TypeID, err == nil)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, map[string]any{"queued": true})
}

// lineLabel resolves a stable metric label for a line (name, else its id).
func (h *Handlers) lineLabel(id int64) string {
	if l, err := h.store.GetLine(id); err == nil && l != nil && l.Name != "" {
		return l.Name
	}
	return strconv.FormatInt(id, 10)
}
