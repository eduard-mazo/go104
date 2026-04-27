package iec104

import (
	"fmt"
	"log/slog"
	"sync"

	"go104/internal/hub"
	"go104/internal/models"
	"go104/internal/store"
)

// Master is a registry of independent LineWorkers.
// Each line (and its TCP connection) is fully isolated.
type Master struct {
	mu      sync.RWMutex
	workers map[int64]*LineWorker
	store   store.Store
	hub     *hub.Hub
}

func NewMaster(s store.Store, h *hub.Hub) *Master {
	return &Master{
		workers: make(map[int64]*LineWorker),
		store:   s,
		hub:     h,
	}
}

// StartAll loads enabled lines from the store and starts each one.
func (m *Master) StartAll() error {
	lines, err := m.store.ListLines()
	if err != nil {
		return fmt.Errorf("list lines: %w", err)
	}
	for _, line := range lines {
		if !line.Enabled {
			continue
		}
		if err := m.StartLine(line); err != nil {
			slog.Error("start line", "id", line.ID, "name", line.Name, "err", err)
		}
	}
	return nil
}

// StopAll signals every worker to stop (does not wait for termination).
func (m *Master) StopAll() {
	m.mu.RLock()
	ws := make([]*LineWorker, 0, len(m.workers))
	for _, w := range m.workers {
		ws = append(ws, w)
	}
	m.mu.RUnlock()
	for _, w := range ws {
		w.Stop()
	}
}

// StartLine creates and starts a LineWorker for the given line.
// Returns an error if the line is already running.
func (m *Master) StartLine(line models.Line) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.workers[line.ID]; ok {
		return fmt.Errorf("line %d already running", line.ID)
	}
	w := NewLineWorker(line, m.store, m.hub)
	if err := w.RefreshSignals(); err != nil {
		return fmt.Errorf("load signals: %w", err)
	}
	w.Start()
	m.workers[line.ID] = w
	slog.Info("line started", "id", line.ID, "name", line.Name,
		"addr", fmt.Sprintf("%s:%d", line.Host, line.Port))
	return nil
}

// StopLine stops the worker for the given line ID.
func (m *Master) StopLine(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.workers[id]
	if !ok {
		return fmt.Errorf("line %d not running", id)
	}
	w.Stop()
	delete(m.workers, id)
	return nil
}

// RestartLine stops (if running) and restarts a line with new config.
func (m *Master) RestartLine(line models.Line) error {
	m.mu.Lock()
	if w, ok := m.workers[line.ID]; ok {
		w.Stop()
		delete(m.workers, line.ID)
	}
	m.mu.Unlock()
	return m.StartLine(line)
}

// SendGI triggers a manual General Interrogation on the given line.
func (m *Master) SendGI(lineID int64) error {
	m.mu.RLock()
	w, ok := m.workers[lineID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("line %d not running", lineID)
	}
	return w.TriggerGI()
}

// SendCommand enqueues a command on the given line.
func (m *Master) SendCommand(cmd models.Command) error {
	m.mu.RLock()
	w, ok := m.workers[cmd.LineID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("line %d not running", cmd.LineID)
	}
	return w.SendCommand(cmd)
}

// RefreshSignals reloads the IOA cache of a running worker.
func (m *Master) RefreshSignals(lineID int64) {
	m.mu.RLock()
	w, ok := m.workers[lineID]
	m.mu.RUnlock()
	if ok {
		if err := w.RefreshSignals(); err != nil {
			slog.Error("refresh signals", "line_id", lineID, "err", err)
		}
	}
}

// IsRunning returns true when the line has an active worker.
func (m *Master) IsRunning(id int64) bool {
	m.mu.RLock()
	_, ok := m.workers[id]
	m.mu.RUnlock()
	return ok
}

// LineState returns the current state of the given line worker.
func (m *Master) LineState(id int64) models.LineState {
	m.mu.RLock()
	w, ok := m.workers[id]
	m.mu.RUnlock()
	if !ok {
		return models.LineStateDisconnected
	}
	return w.State()
}

// LineStats returns rx/tx counters for the given line.
func (m *Master) LineStats(id int64) (rx, tx int64) {
	m.mu.RLock()
	w, ok := m.workers[id]
	m.mu.RUnlock()
	if !ok {
		return 0, 0
	}
	return w.RxCount(), w.TxCount()
}
