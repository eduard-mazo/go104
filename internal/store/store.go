package store

import (
	"database/sql"
	"encoding/json"
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

	ListAllSignals() ([]models.Signal, error)

	InsertHistory(signalID int64, ts float64, value float64, quality uint8) error
	QueryHistory(signalID int64, from, to float64, limit int) ([]models.HistoryPoint, error)

	ListScadaViews() ([]models.ScadaView, error)
	GetScadaView(id int64) (*models.ScadaView, error)
	CreateScadaView(v *models.ScadaView) error
	UpdateScadaView(v *models.ScadaView) error
	DeleteScadaView(id int64) error

	Close() error
}

type sqliteStore struct {
	db       *sql.DB
	mu       sync.Mutex    // single-writer for SQLite (config + periodic snapshot)
	dpCache  sync.Map      // signalID int64 → *models.Datapoint (latest value, RTDB)
	dirty    sync.Map      // signalID int64 → struct{} (changed since last snapshot)
	hist     sync.Map      // signalID int64 → *histBuf (recent trend, in-memory only)
	histCap  int           // ring capacity per signal
	snapStop chan struct{} // stop signal for the snapshot loop
	snapDone chan struct{} // closed when the snapshot loop has exited
}

// New opens (or creates) a SQLite database and runs migrations.
func New(path string) (Store, error) {
	db, err := sql.Open("sqlite", path+"?_journal=WAL&_foreign_keys=on&_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	s := &sqliteStore{
		db: db, histCap: historyPoints(),
		snapStop: make(chan struct{}), snapDone: make(chan struct{}),
	}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := s.loadDPCache(); err != nil {
		return nil, fmt.Errorf("load cache: %w", err)
	}
	go s.snapshotLoop()
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
		`CREATE TABLE IF NOT EXISTS scada_views (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT    NOT NULL,
			width      INTEGER NOT NULL DEFAULT 1400,
			height     INTEGER NOT NULL DEFAULT 900,
			elements   TEXT    NOT NULL DEFAULT '[]',
			created_at TEXT    NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT    NOT NULL DEFAULT (datetime('now'))
		)`,
		// signal_history is intentionally NOT a table — recent trend history lives
		// in an in-memory ring buffer (see history_mem.go). SQLite holds only
		// config + the latest value per signal, not the per-sample stream.
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("stmt failed: %w\n%s", err, stmt[:40])
		}
	}
	// Additive column migrations — ignore error if column already exists.
	s.db.Exec(`ALTER TABLE scada_views ADD COLUMN lines TEXT DEFAULT '[]'`)                   //nolint
	s.db.Exec(`ALTER TABLE lines ADD COLUMN protocol TEXT NOT NULL DEFAULT 'iec104'`)         //nolint
	s.db.Exec(`ALTER TABLE lines ADD COLUMN dnp3_outstation_addr INTEGER NOT NULL DEFAULT 0`) //nolint
	s.db.Exec(`ALTER TABLE lines ADD COLUMN dnp3_master_addr INTEGER NOT NULL DEFAULT 0`)     //nolint
	s.db.Exec(`ALTER TABLE signals ADD COLUMN point_type TEXT NOT NULL DEFAULT ''`)           //nolint
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
		t1_ms,t2_ms,t3_ms,gi_interval_s,protocol,dnp3_outstation_addr,dnp3_master_addr,enabled,created_at FROM lines ORDER BY id`)
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
			&l.K, &l.W, &l.T1MS, &l.T2MS, &l.T3MS, &l.GIIntervalS,
			&l.Protocol, &l.DNP3OutstationAddr, &l.DNP3MasterAddr, &enabled, &created); err != nil {
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
		t1_ms,t2_ms,t3_ms,gi_interval_s,protocol,dnp3_outstation_addr,dnp3_master_addr,enabled,created_at FROM lines WHERE id=?`, id).
		Scan(&l.ID, &l.Name, &l.Host, &l.Port, &ca,
			&l.K, &l.W, &l.T1MS, &l.T2MS, &l.T3MS, &l.GIIntervalS,
			&l.Protocol, &l.DNP3OutstationAddr, &l.DNP3MasterAddr, &enabled, &created)
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
	proto := l.Protocol
	if proto == "" {
		proto = models.ProtocolIEC104
	}
	res, err := s.db.Exec(`INSERT INTO lines (name,host,port,common_address,k,w,t1_ms,t2_ms,t3_ms,gi_interval_s,protocol,dnp3_outstation_addr,dnp3_master_addr,enabled)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		l.Name, l.Host, l.Port, l.CommonAddr, l.K, l.W,
		l.T1MS, l.T2MS, l.T3MS, l.GIIntervalS, proto, l.DNP3OutstationAddr, l.DNP3MasterAddr, boolToInt(l.Enabled))
	if err != nil {
		return err
	}
	l.ID, _ = res.LastInsertId()
	return nil
}

func (s *sqliteStore) UpdateLine(l *models.Line) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	proto := l.Protocol
	if proto == "" {
		proto = models.ProtocolIEC104
	}
	_, err := s.db.Exec(`UPDATE lines SET name=?,host=?,port=?,common_address=?,k=?,w=?,
		t1_ms=?,t2_ms=?,t3_ms=?,gi_interval_s=?,protocol=?,dnp3_outstation_addr=?,dnp3_master_addr=?,enabled=? WHERE id=?`,
		l.Name, l.Host, l.Port, l.CommonAddr, l.K, l.W,
		l.T1MS, l.T2MS, l.T3MS, l.GIIntervalS, proto, l.DNP3OutstationAddr, l.DNP3MasterAddr, boolToInt(l.Enabled), l.ID)
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
		scale,offset_val,description,point_type FROM signals WHERE line_id=? ORDER BY ioa`, lineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sigs []models.Signal
	for rows.Next() {
		var sig models.Signal
		var kind string
		if err := rows.Scan(&sig.ID, &sig.LineID, &sig.Name, &sig.IOA, &sig.TypeID,
			&kind, &sig.Unit, &sig.Scale, &sig.Offset, &sig.Description, &sig.PointType); err != nil {
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
		scale,offset_val,description,point_type FROM signals WHERE id=?`, id).
		Scan(&sig.ID, &sig.LineID, &sig.Name, &sig.IOA, &sig.TypeID,
			&kind, &sig.Unit, &sig.Scale, &sig.Offset, &sig.Description, &sig.PointType)
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
	res, err := s.db.Exec(`INSERT INTO signals (line_id,name,ioa,type_id,signal_type,unit,scale,offset_val,description,point_type)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		sig.LineID, sig.Name, sig.IOA, sig.TypeID, string(sig.Kind),
		sig.Unit, sig.Scale, sig.Offset, sig.Description, sig.PointType)
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
		scale=?,offset_val=?,description=?,point_type=? WHERE id=?`,
		sig.Name, sig.IOA, sig.TypeID, string(sig.Kind),
		sig.Unit, sig.Scale, sig.Offset, sig.Description, sig.PointType, sig.ID)
	return err
}

func (s *sqliteStore) DeleteSignal(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dpCache.Delete(id)
	s.dirty.Delete(id)
	s.hist.Delete(id)
	_, err := s.db.Exec(`DELETE FROM signals WHERE id=?`, id)
	return err
}

// --- Datapoints ---

// UpsertDatapoint updates only the in-memory RTDB (dpCache) and marks the signal
// dirty — it does NOT write SQLite on the hot path. A background snapshot
// (snapshot.go) persists changed current values in batches so a restart still
// shows last-known values via loadDPCache. The live UI is fed from dpCache +
// the WebSocket hub, neither of which needs the per-sample disk write.
func (s *sqliteStore) UpsertDatapoint(dp *models.Datapoint) error {
	clone := *dp
	s.dpCache.Store(dp.SignalID, &clone)
	s.dirty.Store(dp.SignalID, struct{}{})
	return nil
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

// --- All signals (for SCADA signal picker) ---

func (s *sqliteStore) ListAllSignals() ([]models.Signal, error) {
	rows, err := s.db.Query(`SELECT s.id,s.line_id,s.name,s.ioa,s.type_id,s.signal_type,
		s.unit,s.scale,s.offset_val,s.description,s.point_type,l.name
		FROM signals s JOIN lines l ON l.id=s.line_id
		ORDER BY l.name, s.ioa`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sigs []models.Signal
	for rows.Next() {
		var sig models.Signal
		var kind string
		if err := rows.Scan(&sig.ID, &sig.LineID, &sig.Name, &sig.IOA, &sig.TypeID,
			&kind, &sig.Unit, &sig.Scale, &sig.Offset, &sig.Description, &sig.PointType, &sig.LineName); err != nil {
			return nil, err
		}
		sig.Kind = models.SignalKind(kind)
		sigs = append(sigs, sig)
	}
	return sigs, rows.Err()
}

// --- SCADA views ---

func (s *sqliteStore) ListScadaViews() ([]models.ScadaView, error) {
	rows, err := s.db.Query(`SELECT id,name,width,height,updated_at FROM scada_views ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var views []models.ScadaView
	for rows.Next() {
		var v models.ScadaView
		var updated string
		if err := rows.Scan(&v.ID, &v.Name, &v.Width, &v.Height, &updated); err != nil {
			return nil, err
		}
		v.Elements = json.RawMessage("[]")
		v.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		views = append(views, v)
	}
	return views, rows.Err()
}

func (s *sqliteStore) GetScadaView(id int64) (*models.ScadaView, error) {
	v := &models.ScadaView{}
	var created, updated, elements string
	var linesStr sql.NullString
	err := s.db.QueryRow(`SELECT id,name,width,height,elements,COALESCE(lines,'[]'),created_at,updated_at FROM scada_views WHERE id=?`, id).
		Scan(&v.ID, &v.Name, &v.Width, &v.Height, &elements, &linesStr, &created, &updated)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	v.Elements = json.RawMessage(elements)
	if linesStr.Valid && linesStr.String != "" {
		v.Lines = json.RawMessage(linesStr.String)
	} else {
		v.Lines = json.RawMessage("[]")
	}
	v.CreatedAt, _ = time.Parse(time.RFC3339, created)
	v.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return v, nil
}

func (s *sqliteStore) CreateScadaView(v *models.ScadaView) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	elements := string(v.Elements)
	if elements == "" {
		elements = "[]"
	}
	lines := string(v.Lines)
	if lines == "" {
		lines = "[]"
	}
	res, err := s.db.Exec(`INSERT INTO scada_views (name,width,height,elements,lines) VALUES (?,?,?,?,?)`,
		v.Name, v.Width, v.Height, elements, lines)
	if err != nil {
		return err
	}
	v.ID, _ = res.LastInsertId()
	return nil
}

func (s *sqliteStore) UpdateScadaView(v *models.ScadaView) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	elements := string(v.Elements)
	if elements == "" {
		elements = "[]"
	}
	lines := string(v.Lines)
	if lines == "" {
		lines = "[]"
	}
	_, err := s.db.Exec(`UPDATE scada_views SET name=?,width=?,height=?,elements=?,lines=?,updated_at=datetime('now') WHERE id=?`,
		v.Name, v.Width, v.Height, elements, lines, v.ID)
	return err
}

func (s *sqliteStore) DeleteScadaView(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM scada_views WHERE id=?`, id)
	return err
}

// --- Signal history ---
// NOTE: InsertHistory / QueryHistory are implemented in history_mem.go against
// an in-memory ring buffer. They deliberately do NOT touch SQLite — per-sample
// history was the single-writer bottleneck (≈60% of ingest write cost) and is
// not what an edge console's SQLite is for (config + latest value only).

func (s *sqliteStore) Close() error {
	close(s.snapStop)
	<-s.snapDone // wait for the final snapshot flush
	return s.db.Close()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
