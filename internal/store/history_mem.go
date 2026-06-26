package store

import (
	"os"
	"strconv"
	"sync"

	"go104/internal/models"
)

// Recent signal history lives in memory, not SQLite. Per-sample history was the
// single-writer bottleneck (≈60% of ingest write cost in the load test) and an
// edge console only needs a recent trend window for its chart — long-term
// archival is the historian's job (Sparkplug/TimescaleDB upstream). Each signal
// keeps a fixed-size ring of the most recent points; appends are O(1) and never
// touch disk, so ingest no longer serializes on the SQLite writer.
//
// Memory ≈ histCap × ~24 B × (signals that report). Default 2000 points/signal
// ≈ 48 KB/signal; tune with GO104_HISTORY_POINTS (0 disables history entirely).

type histBuf struct {
	mu  sync.Mutex
	buf []models.HistoryPoint
	n   int // total appended (monotonic); position = n % len(buf)
}

// snapshot returns the buffered points in chronological order. Caller holds b.mu.
func (b *histBuf) snapshot() []models.HistoryPoint {
	c := len(b.buf)
	if b.n <= c {
		out := make([]models.HistoryPoint, b.n)
		copy(out, b.buf[:b.n])
		return out
	}
	out := make([]models.HistoryPoint, c)
	start := b.n % c
	for i := range c {
		out[i] = b.buf[(start+i)%c]
	}
	return out
}

func historyPoints() int {
	if v := os.Getenv("GO104_HISTORY_POINTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return 2000
}

// InsertHistory appends one point to the signal's in-memory ring. No SQLite, no
// shared writer lock — only the per-signal buffer's mutex.
func (s *sqliteStore) InsertHistory(signalID int64, ts float64, value float64, quality uint8) error {
	c := s.histCap
	if c <= 0 {
		return nil // history disabled
	}
	v, ok := s.hist.Load(signalID)
	if !ok {
		v, _ = s.hist.LoadOrStore(signalID, &histBuf{buf: make([]models.HistoryPoint, c)})
	}
	b := v.(*histBuf)
	b.mu.Lock()
	b.buf[b.n%c] = models.HistoryPoint{TS: ts, Value: value, Quality: quality}
	b.n++
	b.mu.Unlock()
	return nil
}

// QueryHistory returns buffered points in [from,to] ascending, capped to limit
// (matching the previous SQLite ASC-LIMIT semantics).
func (s *sqliteStore) QueryHistory(signalID int64, from, to float64, limit int) ([]models.HistoryPoint, error) {
	if limit <= 0 || limit > 5000 {
		limit = 2000
	}
	v, ok := s.hist.Load(signalID)
	if !ok {
		return []models.HistoryPoint{}, nil
	}
	b := v.(*histBuf)
	b.mu.Lock()
	all := b.snapshot()
	b.mu.Unlock()

	out := make([]models.HistoryPoint, 0, limit)
	for _, p := range all {
		if p.TS >= from && p.TS <= to {
			out = append(out, p)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
