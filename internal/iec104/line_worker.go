package iec104

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"go104/internal/hub"
	"go104/internal/metrics"
	"go104/internal/models"
	"go104/internal/store"
)

// LineWorker manages one independent IEC104 master connection.
// Each worker runs in its own goroutine tree and has no shared
// counters or sequence numbers with other workers.
type LineWorker struct {
	cfg models.Line
	s   store.Store
	h   *hub.Hub

	mu    sync.RWMutex
	state models.LineState

	rxCount atomic.Int64
	txCount atomic.Int64

	sigMu    sync.RWMutex
	sigCache map[int]*models.Signal // IOA → Signal

	stopCh   chan struct{}
	stopOnce sync.Once
	cmdCh    chan models.Command
	giCh     chan struct{}

	backoffN int // reconnect attempt counter
}

// NewLineWorker creates a worker (does not start it).
func NewLineWorker(cfg models.Line, s store.Store, h *hub.Hub) *LineWorker {
	return &LineWorker{
		cfg:      cfg,
		s:        s,
		h:        h,
		state:    models.LineStateDisconnected,
		sigCache: make(map[int]*models.Signal),
		stopCh:   make(chan struct{}),
		cmdCh:    make(chan models.Command, 16),
		giCh:     make(chan struct{}, 1),
	}
}

// Start launches the worker goroutine.
func (w *LineWorker) Start() {
	go w.run()
}

// Stop signals the worker to disconnect and exit permanently.
func (w *LineWorker) Stop() {
	w.stopOnce.Do(func() { close(w.stopCh) })
}

// TriggerGI requests a manual General Interrogation (non-blocking).
func (w *LineWorker) TriggerGI() error {
	select {
	case w.giCh <- struct{}{}:
		return nil
	default:
		return fmt.Errorf("GI already pending")
	}
}

// SendCommand enqueues a command for transmission.
func (w *LineWorker) SendCommand(cmd models.Command) error {
	select {
	case w.cmdCh <- cmd:
		return nil
	default:
		return fmt.Errorf("command queue full")
	}
}

// RefreshSignals reloads the IOA→Signal cache from the store.
func (w *LineWorker) RefreshSignals() error {
	cache, err := w.s.GetSignalsByLine(w.cfg.ID)
	if err != nil {
		return err
	}
	w.sigMu.Lock()
	w.sigCache = cache
	w.sigMu.Unlock()
	return nil
}

// State returns the current connection state (thread-safe).
func (w *LineWorker) State() models.LineState {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.state
}

// RxCount / TxCount return cumulative frame counts.
func (w *LineWorker) RxCount() int64 { return w.rxCount.Load() }
func (w *LineWorker) TxCount() int64 { return w.txCount.Load() }

// --- internal ---

func (w *LineWorker) setState(s models.LineState) {
	w.mu.Lock()
	w.state = s
	w.mu.Unlock()
	addr := fmt.Sprintf("%s:%d", w.cfg.Host, w.cfg.Port)
	w.h.BroadcastLineStatus(w.cfg.ID, string(s), w.rxCount.Load(), w.txCount.Load(), addr)
	metrics.LineState(w.cfg.Name, "iec104", string(s))
	slog.Info("[LINE] state change", "id", w.cfg.ID, "name", w.cfg.Name, "state", s, "addr", addr)
}

func (w *LineWorker) addr() string {
	return fmt.Sprintf("%s:%d", w.cfg.Host, w.cfg.Port)
}

func (w *LineWorker) backoffDur() time.Duration {
	d := time.Duration(1<<min(w.backoffN, 6)) * time.Second // cap at 64s
	w.backoffN++
	return d
}

func (w *LineWorker) run() {
	defer w.setState(models.LineStateStopped)

	for {
		select {
		case <-w.stopCh:
			return
		default:
		}

		w.setState(models.LineStateConnecting)
		err := w.dialAndRun()
		if err != nil {
			slog.Error("[LINE] disconnected", "id", w.cfg.ID, "addr", w.addr(), "err", err)
		}
		w.setState(models.LineStateDisconnected)
		w.broadcastStaleSignals()

		select {
		case <-w.stopCh:
			return
		case <-time.After(w.backoffDur()):
		}
	}
}

func (w *LineWorker) dialAndRun() error {
	conn, err := net.DialTimeout("tcp", w.addr(), 10*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	// reset backoff on successful TCP connect
	w.backoffN = 0

	recvCh := make(chan *APCI, 64)
	sendCh := make(chan []byte, 128)
	errCh := make(chan error, 2)

	// reader goroutine
	go func() {
		for {
			apci, err := ReadFrame(conn)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
			select {
			case recvCh <- apci:
			case <-w.stopCh:
				return
			}
		}
	}()

	// writer goroutine — exits when sendCh is closed
	go func() {
		for frame := range sendCh {
			if _, err := conn.Write(frame); err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
		}
	}()

	defer close(sendCh)

	// per-connection counters (SSN/RSN isolated per worker invocation)
	var (
		ssn         uint16
		rsn         uint16
		ackRSN      uint16
		unackedSent int
		unackedRcvd int
		active      bool
		giPending   bool // GI sent, waiting for act-term
	)

	send := func(frame []byte) {
		sendCh <- frame
	}
	sendI := func(asduBytes []byte) {
		// Use current SSN first (N(S) starts at 0), then increment.
		send(BuildIFrame(ssn, rsn, asduBytes))
		ssn = (ssn + 1) % 32768
		unackedSent++
		w.txCount.Add(1)
	}
	sendS := func() {
		send(BuildSFrame(rsn))
		unackedRcvd = 0
	}
	ackIncoming := func(newRSN uint16) {
		diff := int(SeqDiff(newRSN, ackRSN))
		if diff > 0 && diff <= unackedSent {
			unackedSent -= diff
			ackRSN = newRSN
		}
	}

	// T1: unack/response timeout
	// T2: delayed S-frame ACK
	// T3: idle TESTFR heartbeat
	t1 := time.NewTimer(time.Duration(w.cfg.T1MS) * time.Millisecond)
	t1.Stop()
	t2 := time.NewTimer(time.Duration(w.cfg.T2MS) * time.Millisecond)
	t2.Stop()
	t3 := time.NewTicker(time.Duration(w.cfg.T3MS) * time.Millisecond)
	defer t1.Stop()
	defer t2.Stop()
	defer t3.Stop()

	startT1 := func() { t1.Reset(time.Duration(w.cfg.T1MS) * time.Millisecond) }
	stopT1 := func() { t1.Stop() }

	var giTick <-chan time.Time
	if w.cfg.GIIntervalS > 0 {
		gt := time.NewTicker(time.Duration(w.cfg.GIIntervalS) * time.Second)
		defer gt.Stop()
		giTick = gt.C
	}

	// Initiate connection with STARTDT
	send(BuildUFrame(STARTDT_ACT))
	startT1()

	for {
		select {
		case <-w.stopCh:
			return nil

		case err := <-errCh:
			return err

		case apci := <-recvCh:
			w.rxCount.Add(1)
			t3.Reset(time.Duration(w.cfg.T3MS) * time.Millisecond)

			switch apci.Type {
			case FrameTypeU:
				switch apci.UFunc {
				case STARTDT_CON:
					active = true
					stopT1()
					w.setState(models.LineStateActive)
					w.sendGI(sendI, &giPending)
					startT1()

				case TESTFR_ACT:
					send(BuildUFrame(TESTFR_CON))

				case TESTFR_CON:
					if unackedSent == 0 {
						stopT1()
					}

				case STOPDT_CON:
					return nil
				}

			case FrameTypeS:
				ackIncoming(apci.RSN)
				if unackedSent == 0 {
					stopT1()
				}

			case FrameTypeI:
				if !active {
					break
				}
				rsn = (apci.SSN + 1) % 32768
				ackIncoming(apci.RSN)
				if unackedSent == 0 {
					stopT1()
				}

				// W-window: send S-frame after W received I-frames
				unackedRcvd++
				if unackedRcvd >= w.cfg.W {
					sendS()
					t2.Stop()
				} else {
					t2.Reset(time.Duration(w.cfg.T2MS) * time.Millisecond)
				}

				if len(apci.ASDU) > 0 {
					if err := w.processASDU(apci.ASDU, &giPending); err != nil {
						slog.Warn("[LINE] ASDU error", "id", w.cfg.ID, "err", err)
					}
				}
			}

		case <-t1.C:
			return fmt.Errorf("T1 timeout (%dms)", w.cfg.T1MS)

		case <-t2.C:
			sendS()

		case <-t3.C:
			send(BuildUFrame(TESTFR_ACT))
			// Only start T1 for TESTFR if no I-frames are pending;
			// if unackedSent > 0, T1 is already running for those.
			if unackedSent == 0 {
				startT1()
			}

		case <-giTick:
			if active && !giPending {
				w.sendGI(sendI, &giPending)
				startT1()
			}

		case <-w.giCh:
			if active {
				// Manual GI: allow even if one is pending (operator intent)
				giPending = false
				w.sendGI(sendI, &giPending)
				startT1()
			}

		case cmd := <-w.cmdCh:
			if active {
				if b := BuildCommandAsdu(cmd, uint16(w.cfg.CommonAddr)); len(b) > 0 {
					sendI(b)
					startT1()
				}
			}
		}
	}
}

// sendGI sends a General Interrogation I-frame and marks giPending.
func (w *LineWorker) sendGI(sendI func([]byte), giPending *bool) {
	slog.Info("[LINE] sending GI", "id", w.cfg.ID, "ca", w.cfg.CommonAddr)
	sendI(BuildGIAsdu(uint16(w.cfg.CommonAddr)))
	*giPending = true
	w.h.BroadcastRaw(map[string]any{
		"type":    "gi_start",
		"line_id": w.cfg.ID,
	})
}

// broadcastStaleSignals notifies the hub that all cached signals on this line
// are stale (connection lost) so the UI can mark them as invalid.
func (w *LineWorker) broadcastStaleSignals() {
	w.sigMu.RLock()
	defer w.sigMu.RUnlock()
	for _, sig := range w.sigCache {
		w.h.BroadcastRaw(map[string]any{
			"type":      "signal_stale",
			"signal_id": sig.ID,
			"line_id":   w.cfg.ID,
		})
	}
}

// processASDU decodes an ASDU, handles GI lifecycle, stores and broadcasts data objects.
func (w *LineWorker) processASDU(data []byte, giPending *bool) error {
	asdu, err := ParseASDU(data)
	if err != nil {
		return err
	}

	switch asdu.COT {
	case CotActCon:
		// Activation confirmation — log GI confirm, skip data processing
		if int(asdu.TypeID) == C_IC_NA_1 {
			slog.Info("[LINE] GI confirmed by slave", "id", w.cfg.ID, "ca", asdu.CA)
		}
		return nil

	case CotActTerm:
		if int(asdu.TypeID) == C_IC_NA_1 {
			*giPending = false
			slog.Info("[LINE] GI complete", "id", w.cfg.ID)
			w.h.BroadcastRaw(map[string]any{
				"type":    "gi_complete",
				"line_id": w.cfg.ID,
			})
		}
		return nil

	case CotSpontaneous, CotRequest, CotInterrogated, CotPeriodic, CotBackground:
		// data — process below
	default:
		return nil
	}

	if !IsMonitoringType(int(asdu.TypeID)) {
		return nil
	}

	now := time.Now()
	w.sigMu.RLock()
	defer w.sigMu.RUnlock()

	for _, obj := range asdu.Objects {
		sig, ok := w.sigCache[int(obj.IOA)]
		if !ok {
			continue
		}

		ts := now
		if obj.HasTime {
			ts = obj.Timestamp
		}

		value := obj.Value
		if sig.Kind == models.SignalKindAnalog {
			value = obj.Value*sig.Scale + sig.Offset
		}

		dp := &models.Datapoint{
			SignalID:   sig.ID,
			LineID:     w.cfg.ID,
			IOA:        sig.IOA,
			Name:       sig.Name,
			Kind:       sig.Kind,
			TypeID:     int(asdu.TypeID),
			Unit:       sig.Unit,
			Value:      value,
			RawValue:   obj.Value,
			Quality:    obj.Quality,
			Timestamp:  ts,
			ReceivedAt: now,
		}

		if err := w.s.UpsertDatapoint(dp); err != nil {
			slog.Error("[LINE] upsert datapoint", "id", w.cfg.ID, "signal", sig.ID, "err", err)
		}
		tsEpoch := float64(dp.Timestamp.UnixNano()) / 1e9
		if err := w.s.InsertHistory(sig.ID, tsEpoch, dp.Value, dp.Quality); err != nil {
			slog.Warn("[LINE] insert history", "signal", sig.ID, "err", err)
		}
		w.h.BroadcastDatapoint(dp)
		metrics.IngestSample(w.cfg.Name, "iec104", dp.QualityOK())
	}
	return nil
}
