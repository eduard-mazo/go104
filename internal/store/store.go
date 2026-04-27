package store

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"go104/internal/models"
	_ "modernc.org/sqlite"
)

// Store defines all persistence operations.
type Store interface {
	ListLines() ([]models.Line, error)
	GetLine(id int64) (*models.Line, error)
	CreateLine(line *models.Line) error
	UpdateLine(line *models.Line) error
	DeleteLine(id int64) error

	ListSignals(lineID int64) ([]models.Signal, error)
	GetSignal(id int64) (*models.Signal, error)
	GetSignalsByLine(lineID int64) (map[int]*models.Signal, error)
	CreateSignal(sig *models.Signal) error
	UpdateSignal(sig *models.Signal) error
	DeleteSignal(id int64) error

	UpsertDatapoint(dp *models.Datapoint) error
	ListDatapoints(lineID int64) ([]models.Datapoint, error)

	Close() error
}

type sqliteStore struct {
	db      *sql.DB
	mu      sync.Mutex  // single-writer for SQLite
	dpCache sync.Map    // signalID int64 → *models.Datapoint
}

// New opens (or creates) a SQLite database and runs migrations.
func New(path string) (Store, error) {
	db, err := sql.Open("sqlite", path+"?_journal=WAL&_foreign_keys=on&_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	s := &sqliteStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := s.loadDPCache(); err != nil {
		return nil, fmt.Errorf("load cache: %w", err)
	}
	return s, nil
}

func (s *sqliteStore) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS lines (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			name            TEXT    NOT NULL,
			host            TEXT    NOT NULL,
			port            INTEGER NOT NULL,
			common_address  INTEGER NOT NULL DEFAULT 1,
			k               INTEGER NOT NULL DEFAULT 12,
			w               INTEGER NOT NULL DEFAULT 8,
			t1_ms           INTEGER NOT NULL DEFAULT 15000,
			t2_ms           INTEGER NOT NULL DEFAULT 10000,
			t3_ms           INTEGER NOT NULL DEFAULT 20000,
			gi_interval_s   INTEGER NOT NULL DEFAULT 900,
			enabled         INTEGER NOT NULL DEFAULT 1,
			created_at      TEXT    NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS signals (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			line_id     INTEGER NOT NULL REFERENCES lines(id) ON DELETE CASCADE,
			name        TEXT    NOT NULL,
			ioa         INTEGER NOT NULL,
			type_id     INTEGER NOT NULL,
			signal_type TEXT    NOT NULL DEFAULT 'analog',
			unit        TEXT    NOT NULL DEFAULT '',
			scale       REAL    NOT NULL DEFAULT 1.0,
			offset_val  REAL    NOT NULL DEFAULT 0.0,
			description TEXT    NOT NULL DEFAULT '',
			UNIQUE(line_id, ioa)
		)`,
		`CREATE TABLE IF NOT EXISTS datapoints (
			signal_id   INTEGER PRIMARY KEY REFERENCES signals(id) ON DELETE CASCADE,
			line_id     INTEGER NOT NULL,
			ioa         INTEGER NOT NULL,
			name        TEXT    NOT NULL,
			signal_type TEXT    NOT NULL,
			type_id     INTEGER NOT NULL,
			unit        TEXT    NOT NULL DEFAULT '',
			value       REAL    NOT NULL DEFAULT 0,
			raw_value   REAL    NOT NULL DEFAULT 0,
			quality     INTEGER NOT NULL DEFAULT 128,
			timestamp   TEXT    NOT NULL,
			received_at TEXT    NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_signals_line ON signals(line_id)`,
		`CREATE INDEX IF NOT EXISTS idx_datapoints_line ON datapoints(line_id)`,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("stmt failed: %w\n%s", err, stmt[:40])
		}
	}
	return nil
}

func (s *sqliteStore) loadDPCache() error {
	rows, err := s.db.Query(`SELECT signal_id,line_id,ioa,name,signal_type,type_id,unit,
		value,raw_value,quality,timestamp,received_at FROM datapoints`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		dp := &models.Datapoint{}
		var ts, ra, kind string
		if err := rows.Scan(&dp.SignalID, &dp.LineID, &dp.IOA, &dp.Name, &kind,
			&dp.TypeID, &dp.Unit, &dp.Value, &dp.RawValue, &dp.Quality, &ts, &ra); err != nil {
			return err
		}
		dp.Kind = models.SignalKind(kind)
		dp.Timestamp, _ = time.Parse(time.RFC3339Nano, ts)
		dp.ReceivedAt, _ = time.Parse(time.RFC3339Nano, ra)
		s.dpCache.Store(dp.SignalID, dp)
	}
	return rows.Err()
}

// --- Lines ---

func (s *sqliteStore) ListLines() ([]models.Line, error) {
	rows, err := s.db.Query(`SELECT id,name,host,port,common_address,k,w,
		t1_ms,t2_ms,t3_ms,gi_interval_s,enabled,created_at FROM lines ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lines []models.Line
	for rows.Next() {
		var l models.Line
		var ca, enabled int
		var created string
		if err := rows.Scan(&l.ID, &l.Name, &l.Host, &l.Port, &ca,
			&l.K, &l.W, &l.T1MS, &l.T2MS, &l.T3MS, &l.GIIntervalS, &enabled, &created); err != nil {
			return nil, err
		}
		l.CommonAddr = ca
		l.Enabled = enabled != 0
		l.CreatedAt, _ = time.Parse(time.RFC3339, created)
		lines = append(lines, l)
	}
	return lines, rows.Err()
}

func (s *sqliteStore) GetLine(id int64) (*models.Line, error) {
	l := &models.Line{}
	var ca, enabled int
	var created string
	err := s.db.QueryRow(`SELECT id,name,host,port,common_address,k,w,
		t1_ms,t2_ms,t3_ms,gi_interval_s,enabled,created_at FROM lines WHERE id=?`, id).
		Scan(&l.ID, &l.Name, &l.Host, &l.Port, &ca,
			&l.K, &l.W, &l.T1MS, &l.T2MS, &l.T3MS, &l.GIIntervalS, &enabled, &created)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	l.CommonAddr = ca
	l.Enabled = enabled != 0
	l.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return l, nil
}

func (s *sqliteStore) CreateLine(l *models.Line) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	res, err := s.db.Exec(`INSERT INTO lines (name,host,port,common_address,k,w,t1_ms,t2_ms,t3_ms,gi_interval_s,enabled)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		l.Name, l.Host, l.Port, l.CommonAddr, l.K, l.W,
		l.T1MS, l.T2MS, l.T3MS, l.GIIntervalS, boolToInt(l.Enabled))
	if err != nil {
		return err
	}
	l.ID, _ = res.LastInsertId()
	return nil
}

func (s *sqliteStore) UpdateLine(l *models.Line) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE lines SET name=?,host=?,port=?,common_address=?,k=?,w=?,
		t1_ms=?,t2_ms=?,t3_ms=?,gi_interval_s=?,enabled=? WHERE id=?`,
		l.Name, l.Host, l.Port, l.CommonAddr, l.K, l.W,
		l.T1MS, l.T2MS, l.T3MS, l.GIIntervalS, boolToInt(l.Enabled), l.ID)
	return err
}

func (s *sqliteStore) DeleteLine(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM lines WHERE id=?`, id)
	return err
}

// --- Signals ---

func (s *sqliteStore) ListSignals(lineID int64) ([]models.Signal, error) {
	rows, err := s.db.Query(`SELECT id,line_id,name,ioa,type_id,signal_type,unit,
		scale,offset_val,description FROM signals WHERE line_id=? ORDER BY ioa`, lineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sigs []models.Signal
	for rows.Next() {
		var sig models.Signal
		var kind string
		if err := rows.Scan(&sig.ID, &sig.LineID, &sig.Name, &sig.IOA, &sig.TypeID,
			&kind, &sig.Unit, &sig.Scale, &sig.Offset, &sig.Description); err != nil {
			return nil, err
		}
		sig.Kind = models.SignalKind(kind)
		sigs = append(sigs, sig)
	}
	return sigs, rows.Err()
}

func (s *sqliteStore) GetSignal(id int64) (*models.Signal, error) {
	sig := &models.Signal{}
	var kind string
	err := s.db.QueryRow(`SELECT id,line_id,name,ioa,type_id,signal_type,unit,
		scale,offset_val,description FROM signals WHERE id=?`, id).
		Scan(&sig.ID, &sig.LineID, &sig.Name, &sig.IOA, &sig.TypeID,
			&kind, &sig.Unit, &sig.Scale, &sig.Offset, &sig.Description)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sig.Kind = models.SignalKind(kind)
	return sig, nil
}

func (s *sqliteStore) GetSignalsByLine(lineID int64) (map[int]*models.Signal, error) {
	sigs, err := s.ListSignals(lineID)
	if err != nil {
		return nil, err
	}
	m := make(map[int]*models.Signal, len(sigs))
	for i := range sigs {
		m[sigs[i].IOA] = &sigs[i]
	}
	return m, nil
}

func (s *sqliteStore) CreateSignal(sig *models.Signal) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	res, err := s.db.Exec(`INSERT INTO signals (line_id,name,ioa,type_id,signal_type,unit,scale,offset_val,description)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		sig.LineID, sig.Name, sig.IOA, sig.TypeID, string(sig.Kind),
		sig.Unit, sig.Scale, sig.Offset, sig.Description)
	if err != nil {
		return err
	}
	sig.ID, _ = res.LastInsertId()
	return nil
}

func (s *sqliteStore) UpdateSignal(sig *models.Signal) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE signals SET name=?,ioa=?,type_id=?,signal_type=?,unit=?,
		scale=?,offset_val=?,description=? WHERE id=?`,
		sig.Name, sig.IOA, sig.TypeID, string(sig.Kind),
		sig.Unit, sig.Scale, sig.Offset, sig.Description, sig.ID)
	return err
}

func (s *sqliteStore) DeleteSignal(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dpCache.Delete(id)
	_, err := s.db.Exec(`DELETE FROM signals WHERE id=?`, id)
	return err
}

// --- Datapoints ---

func (s *sqliteStore) UpsertDatapoint(dp *models.Datapoint) error {
	s.mu.Lock()
	_, err := s.db.Exec(`INSERT INTO datapoints
		(signal_id,line_id,ioa,name,signal_type,type_id,unit,value,raw_value,quality,timestamp,received_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(signal_id) DO UPDATE SET
		line_id=excluded.line_id,ioa=excluded.ioa,name=excluded.name,
		signal_type=excluded.signal_type,type_id=excluded.type_id,unit=excluded.unit,
		value=excluded.value,raw_value=excluded.raw_value,quality=excluded.quality,
		timestamp=excluded.timestamp,received_at=excluded.received_at`,
		dp.SignalID, dp.LineID, dp.IOA, dp.Name, string(dp.Kind), dp.TypeID, dp.Unit,
		dp.Value, dp.RawValue, dp.Quality,
		dp.Timestamp.UTC().Format(time.RFC3339Nano),
		dp.ReceivedAt.UTC().Format(time.RFC3339Nano),
	)
	s.mu.Unlock()
	if err == nil {
		clone := *dp
		s.dpCache.Store(dp.SignalID, &clone)
	}
	return err
}

func (s *sqliteStore) ListDatapoints(lineID int64) ([]models.Datapoint, error) {
	sigs, err := s.ListSignals(lineID)
	if err != nil {
		return nil, err
	}
	result := make([]models.Datapoint, 0, len(sigs))
	for _, sig := range sigs {
		if v, ok := s.dpCache.Load(sig.ID); ok {
			result = append(result, *v.(*models.Datapoint))
		} else {
			result = append(result, models.Datapoint{
				SignalID: sig.ID,
				LineID:   lineID,
				IOA:      sig.IOA,
				Name:     sig.Name,
				Kind:     sig.Kind,
				TypeID:   sig.TypeID,
				Unit:     sig.Unit,
				Quality:  0x80, // invalid — no data yet
			})
		}
	}
	return result, nil
}

func (s *sqliteStore) Close() error {
	return s.db.Close()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
