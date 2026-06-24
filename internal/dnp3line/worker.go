// Package dnp3line drives a DNP3 master association for go104: it polls a remote
// outstation through the shared goDnp3 binding and records each measurement as a
// Datapoint, mirroring the IEC-104 LineWorker's store/hub write path so the rest
// of go104 (API, history, SCADA views) treats DNP3 lines the same as IEC-104.
//
// Monitor-only: SendCommand is rejected (DNP3 controls are a later phase).
//
// No build tags here — goDnp3 selects its real opendnp3 binding under
// -tags dnp3_ffi and a pure-Go stub otherwise, so go104's default build stays
// pure-Go (a DNP3 line then connects to nothing) and the dnp3_ffi build polls
// for real.
package dnp3line

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	godnp3 "goDnp3"

	"go104/internal/hub"
	"go104/internal/models"
	"go104/internal/store"
)

// Worker drives one DNP3 master association. It implements both goDnp3.Handler
// (measurement/status/log callbacks) and the line-worker contract the iec104
// Master registry expects (Start/Stop/State/RxCount/TxCount/RefreshSignals/
// TriggerGI/SendCommand).
type Worker struct {
	cfg  models.Line
	s    store.Store
	h    *hub.Hub
	osID string // outstation id for AddOutstation / IntegrityPoll

	mu    sync.RWMutex
	state models.LineState

	rxCount atomic.Int64

	sigMu    sync.RWMutex
	sigCache map[pointKey]*models.Signal // (pointType, index) → Signal

	master godnp3.Master
	cancel context.CancelFunc
}

// pointKey identifies a DNP3 point by object group + index (IOA reused as index).
type pointKey struct {
	pt  string
	idx uint16
}

// NewWorker creates a DNP3 line worker (does not start it).
func NewWorker(cfg models.Line, s store.Store, h *hub.Hub) *Worker {
	return &Worker{
		cfg:      cfg,
		s:        s,
		h:        h,
		osID:     fmt.Sprintf("line-%d", cfg.ID),
		state:    models.LineStateDisconnected,
		sigCache: make(map[pointKey]*models.Signal),
	}
}

// Start brings up the DNP3 master and begins polling.
func (w *Worker) Start() {
	w.setState(models.LineStateConnecting)
	m := godnp3.NewMaster(w)
	if err := m.AddOutstation(w.outstationConfig()); err != nil {
		slog.Error("[DNP3] add outstation", "line", w.cfg.ID, "err", err)
		w.setState(models.LineStateDisconnected)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.master = m
	if err := m.Start(ctx); err != nil {
		slog.Error("[DNP3] master start", "line", w.cfg.ID, "err", err)
		w.setState(models.LineStateDisconnected)
		return
	}
	slog.Info("[DNP3] line started", "id", w.cfg.ID, "addr", w.addr(),
		"outstation", w.cfg.DNP3OutstationAddr, "master", w.cfg.DNP3MasterAddr)
}

// Stop tears down the DNP3 master.
func (w *Worker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	if w.master != nil {
		w.master.Stop()
	}
	w.setState(models.LineStateStopped)
}

func (w *Worker) State() models.LineState {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.state
}

func (w *Worker) RxCount() int64 { return w.rxCount.Load() }
func (w *Worker) TxCount() int64 { return 0 } // polls are issued internally by the lib

// RefreshSignals reloads the (pointType,index)→Signal cache from the store.
func (w *Worker) RefreshSignals() error {
	sigs, err := w.s.ListSignals(w.cfg.ID)
	if err != nil {
		return err
	}
	cache := make(map[pointKey]*models.Signal, len(sigs))
	for i := range sigs {
		sig := &sigs[i]
		pt := sig.PointType
		if pt == "" {
			pt = string(godnp3.PointAnalog) // untyped signal → analog group
		}
		cache[pointKey{pt, uint16(sig.IOA)}] = sig
	}
	w.sigMu.Lock()
	w.sigCache = cache
	w.sigMu.Unlock()
	return nil
}

// TriggerGI maps a manual General Interrogation to a DNP3 integrity poll.
func (w *Worker) TriggerGI() error {
	if w.master == nil {
		return fmt.Errorf("line not running")
	}
	return w.master.IntegrityPoll(w.osID)
}

// SendCommand issues a DNP3 control to the outstation via DirectOperate. The
// command's TypeID (the same IEC-104 codes the API/UI use) picks the control
// kind: single/double commands (C_SC_NA_1=45 / C_DC_NA_1=46) operate a binary
// output (CROB on/off from Value != 0); setpoints (C_SE_NA_1=48 / C_SE_NB_1=49 /
// C_SE_NC_1=50) operate an analog output (Value). IOA is the DNP3 point index.
// Select is ignored — goDnp3 exposes DirectOperate only.
func (w *Worker) SendCommand(cmd models.Command) error {
	if w.master == nil {
		return fmt.Errorf("line not running")
	}
	idx := uint16(cmd.IOA)
	switch cmd.TypeID {
	case 45, 46: // C_SC_NA_1 / C_DC_NA_1 — binary command → CROB
		return w.master.OperateBinary(w.osID, idx, cmd.Value != 0)
	case 48, 49, 50: // C_SE_NA_1 / C_SE_NB_1 / C_SE_NC_1 — setpoint → analog output
		return w.master.OperateAnalog(w.osID, idx, cmd.Value)
	default:
		return fmt.Errorf("DNP3: unsupported control type %d (use 45/46 binary or 48-50 setpoint)", cmd.TypeID)
	}
}

// --- goDnp3.Handler ---

// OnMeasurement records one measurement as a Datapoint (same write path as the
// IEC-104 worker): resolve the signal by (point type, index), apply scale/offset
// for analog kinds, and upsert + history + broadcast.
func (w *Worker) OnMeasurement(m godnp3.Measurement) {
	w.rxCount.Add(1)
	w.sigMu.RLock()
	sig, ok := w.sigCache[pointKey{string(m.PointType), m.Index}]
	w.sigMu.RUnlock()
	if !ok {
		return
	}

	now := time.Now()
	raw := measValue(m)
	value := raw
	if sig.Kind == models.SignalKindAnalog {
		value = raw*sig.Scale + sig.Offset
	}
	ts := m.Time
	if ts.IsZero() {
		ts = now
	}

	dp := &models.Datapoint{
		SignalID:   sig.ID,
		LineID:     w.cfg.ID,
		IOA:        sig.IOA,
		Name:       sig.Name,
		Kind:       sig.Kind,
		TypeID:     sig.TypeID,
		Unit:       sig.Unit,
		Value:      value,
		RawValue:   raw,
		Quality:    iecQuality(m.Quality),
		Timestamp:  ts,
		ReceivedAt: now,
	}
	if err := w.s.UpsertDatapoint(dp); err != nil {
		slog.Error("[DNP3] upsert datapoint", "line", w.cfg.ID, "signal", sig.ID, "err", err)
		return
	}
	tsEpoch := float64(dp.Timestamp.UnixNano()) / 1e9
	if err := w.s.InsertHistory(sig.ID, tsEpoch, dp.Value, dp.Quality); err != nil {
		slog.Warn("[DNP3] insert history", "signal", sig.ID, "err", err)
	}
	w.h.BroadcastDatapoint(dp)
}

func (w *Worker) OnStatusChange(s godnp3.Status) {
	if s.Connected {
		w.setState(models.LineStateActive)
	} else {
		w.setState(models.LineStateDisconnected)
	}
}

func (w *Worker) OnLog(level, msg string) {
	switch level {
	case "error", "warn":
		slog.Warn("[DNP3] "+msg, "line", w.cfg.ID)
	default:
		slog.Debug("[DNP3] "+msg, "line", w.cfg.ID)
	}
}

// --- internal ---

func (w *Worker) outstationConfig() godnp3.OutstationConfig {
	return godnp3.OutstationConfig{
		ID:                w.osID,
		Label:             w.cfg.Name,
		Host:              w.cfg.Host,
		Port:              w.cfg.Port,
		MasterAddress:     uint16(w.cfg.DNP3MasterAddr),
		OutstationAddress: uint16(w.cfg.DNP3OutstationAddr),
		StartupIntegrity:  true,
		IntegrityScanMs:   w.cfg.GIIntervalS * 1000, // periodic integrity from GI interval
		Class1ScanMs:      1000,                     // poll class-1 events so live changes flow
	}
}

func (w *Worker) addr() string {
	return fmt.Sprintf("%s:%d", w.cfg.Host, w.cfg.Port)
}

func (w *Worker) setState(s models.LineState) {
	w.mu.Lock()
	w.state = s
	w.mu.Unlock()
	w.h.BroadcastLineStatus(w.cfg.ID, string(s), w.rxCount.Load(), 0, w.addr())
	slog.Info("[DNP3] state change", "id", w.cfg.ID, "name", w.cfg.Name, "state", s, "addr", w.addr())
}

// measValue extracts the numeric value of a measurement (bools → 0/1).
func measValue(m godnp3.Measurement) float64 {
	switch m.PointType {
	case godnp3.PointBinary, godnp3.PointBinaryOutputStatus:
		if m.BoolValue {
			return 1
		}
		return 0
	case godnp3.PointDoubleBitBinary:
		return float64(m.DBBValue)
	case godnp3.PointCounter, godnp3.PointFrozenCounter:
		return float64(m.UintValue)
	case godnp3.PointAnalog, godnp3.PointAnalogOutputStatus:
		return m.FloatValue
	}
	return 0
}

// iecQuality maps a DNP3 quality flag byte to the IEC-104-style quality byte
// go104 stores: good → 0 (valid), otherwise IV (invalid, 0x80). go104's
// Datapoint.QualityOK() checks the IV/BL bits.
func iecQuality(q godnp3.Quality) uint8 {
	if q.Good() {
		return 0
	}
	return 0x80
}
