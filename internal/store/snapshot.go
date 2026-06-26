package store

import (
	"os"
	"strconv"
	"time"

	"go104/internal/models"
)

// Current values (latest datapoint per signal) live in the in-memory RTDB
// (dpCache). To keep them off the SQLite hot path, UpsertDatapoint never writes
// disk — instead a background loop persists only the *changed* current values in
// one batched transaction every snapshotInterval (and once more on shutdown), so
// a restart still shows last-known values via loadDPCache. This turns N
// per-sample writes/s into one small transaction every interval.

func snapshotInterval() time.Duration {
	if v := os.Getenv("GO104_SNAPSHOT_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 30 * time.Second
}

func (s *sqliteStore) snapshotLoop() {
	defer close(s.snapDone)
	t := time.NewTicker(snapshotInterval())
	defer t.Stop()
	for {
		select {
		case <-s.snapStop:
			s.flushCurrent() // final flush before Close()
			return
		case <-t.C:
			s.flushCurrent()
		}
	}
}

// flushCurrent persists the changed current values to SQLite in a single
// transaction. Off the hot path; contends with config writes only.
func (s *sqliteStore) flushCurrent() {
	var ids []int64
	s.dirty.Range(func(k, _ any) bool {
		ids = append(ids, k.(int64))
		s.dirty.Delete(k)
		return true
	})
	if len(ids) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		s.remarkDirty(ids)
		return
	}
	stmt, err := tx.Prepare(`INSERT INTO datapoints
		(signal_id,line_id,ioa,name,signal_type,type_id,unit,value,raw_value,quality,timestamp,received_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(signal_id) DO UPDATE SET
		line_id=excluded.line_id,ioa=excluded.ioa,name=excluded.name,
		signal_type=excluded.signal_type,type_id=excluded.type_id,unit=excluded.unit,
		value=excluded.value,raw_value=excluded.raw_value,quality=excluded.quality,
		timestamp=excluded.timestamp,received_at=excluded.received_at`)
	if err != nil {
		_ = tx.Rollback()
		s.remarkDirty(ids)
		return
	}
	defer stmt.Close()

	for _, id := range ids {
		v, ok := s.dpCache.Load(id)
		if !ok {
			continue // signal deleted since it was marked dirty
		}
		dp := v.(*models.Datapoint)
		if _, err := stmt.Exec(dp.SignalID, dp.LineID, dp.IOA, dp.Name, string(dp.Kind),
			dp.TypeID, dp.Unit, dp.Value, dp.RawValue, dp.Quality,
			dp.Timestamp.UTC().Format(time.RFC3339Nano),
			dp.ReceivedAt.UTC().Format(time.RFC3339Nano)); err != nil {
			s.dirty.Store(id, struct{}{}) // retry next round
		}
	}
	if err := tx.Commit(); err != nil {
		s.remarkDirty(ids)
	}
}

func (s *sqliteStore) remarkDirty(ids []int64) {
	for _, id := range ids {
		s.dirty.Store(id, struct{}{})
	}
}
