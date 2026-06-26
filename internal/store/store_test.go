package store

import (
	"sync"
	"testing"
	"time"

	"go104/internal/models"
)

func newTestStore(t *testing.T) *sqliteStore {
	t.Helper()
	st, err := New(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st.(*sqliteStore)
}

func seedSignal(t *testing.T, st *sqliteStore) *models.Signal {
	t.Helper()
	l := &models.Line{Name: "l", Host: "127.0.0.1", Port: 20000, Protocol: "dnp3"}
	if err := st.CreateLine(l); err != nil {
		t.Fatal(err)
	}
	sig := &models.Signal{LineID: l.ID, Name: "s", IOA: 1, Kind: models.SignalKindAnalog, PointType: "analog", Scale: 1}
	if err := st.CreateSignal(sig); err != nil {
		t.Fatal(err)
	}
	return sig
}

func mkDP(sig *models.Signal, v float64) *models.Datapoint {
	now := time.Now()
	return &models.Datapoint{SignalID: sig.ID, LineID: sig.LineID, IOA: sig.IOA,
		Name: sig.Name, Kind: sig.Kind, Value: v, Timestamp: now, ReceivedAt: now}
}

// UpsertDatapoint must update the in-memory RTDB but NOT write SQLite on the hot path.
func TestUpsertStaysOffSQLite(t *testing.T) {
	st := newTestStore(t)
	sig := seedSignal(t, st)
	if err := st.UpsertDatapoint(mkDP(sig, 42)); err != nil {
		t.Fatal(err)
	}
	dps, _ := st.ListDatapoints(sig.LineID)
	if len(dps) != 1 || dps[0].Value != 42 {
		t.Fatalf("RTDB not updated: %+v", dps)
	}
	var n int
	if err := st.db.QueryRow(`SELECT count(*) FROM datapoints`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("hot path wrote SQLite: %d rows (expected 0 until snapshot)", n)
	}
}

// A snapshot must persist current values so a fresh store on the same file
// restores last-known values via loadDPCache.
func TestSnapshotPersistsForRestart(t *testing.T) {
	path := t.TempDir() + "/t.db"
	st1, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	ss := st1.(*sqliteStore)
	sig := seedSignal(t, ss)
	if err := ss.UpsertDatapoint(mkDP(sig, 7)); err != nil {
		t.Fatal(err)
	}
	ss.flushCurrent() // force the batched snapshot
	var n int
	ss.db.QueryRow(`SELECT count(*) FROM datapoints`).Scan(&n)
	if n != 1 {
		t.Fatalf("snapshot did not persist: %d rows", n)
	}
	st1.Close()

	st2, err := New(path) // simulate restart
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()
	dps, _ := st2.ListDatapoints(sig.LineID)
	if len(dps) != 1 || dps[0].Value != 7 {
		t.Fatalf("restart lost current value: %+v", dps)
	}
}

// The history ring is bounded and keeps the most-recent window.
func TestHistoryRingBounded(t *testing.T) {
	t.Setenv("GO104_HISTORY_POINTS", "5")
	st := newTestStore(t)
	sig := seedSignal(t, st)
	for i := range 20 {
		st.InsertHistory(sig.ID, float64(i), float64(i), 0)
	}
	pts, _ := st.QueryHistory(sig.ID, 0, 1e18, 100)
	if len(pts) != 5 {
		t.Fatalf("ring not bounded: got %d, want 5", len(pts))
	}
	if pts[0].TS != 15 || pts[4].TS != 19 {
		t.Fatalf("ring kept wrong window: %.0f..%.0f, want 15..19", pts[0].TS, pts[4].TS)
	}
}

// Hammer the hot path concurrently with snapshots + queries (run with -race).
func TestConcurrentHotPath(t *testing.T) {
	st := newTestStore(t)
	sig := seedSignal(t, st)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 2000 {
				st.UpsertDatapoint(mkDP(sig, float64(i)))
				st.InsertHistory(sig.ID, float64(i), float64(i), 0)
			}
		}()
	}
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				st.flushCurrent()
				st.QueryHistory(sig.ID, 0, 1e18, 10)
				time.Sleep(time.Millisecond)
			}
		}
	}()
	wg.Wait()
	close(done)
}
