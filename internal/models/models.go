package models

import (
	"encoding/json"
	"time"
)

type LineState string

const (
	LineStateDisconnected LineState = "DISCONNECTED"
	LineStateConnecting   LineState = "CONNECTING"
	LineStateActive       LineState = "ACTIVE"
	LineStateStopped      LineState = "STOPPED"
)

// Line protocol drivers.
const (
	ProtocolIEC104 = "iec104" // poll an IEC-104 server (default)
	ProtocolDNP3   = "dnp3"   // poll a DNP3 outstation as a master
)

type Line struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	CommonAddr  int    `json:"common_address"`
	K           int    `json:"k"`             // max unacked sent I-frames
	W           int    `json:"w"`             // send S-frame after W received I-frames
	T1MS        int    `json:"t1_ms"`         // unack timeout ms
	T2MS        int    `json:"t2_ms"`         // delayed ACK timeout ms
	T3MS        int    `json:"t3_ms"`         // TESTFR idle interval ms
	GIIntervalS int    `json:"gi_interval_s"` // 0 = disabled (for DNP3: integrity-poll interval)

	// Protocol selects the line driver: "iec104" (default) or "dnp3". The DNP3
	// fields below apply only when Protocol == "dnp3"; Host/Port and GIIntervalS
	// are shared (GIIntervalS becomes the integrity-poll interval).
	Protocol           string `json:"protocol"`
	DNP3OutstationAddr int    `json:"dnp3_outstation_addr"` // remote outstation link addr (typical 1024+)
	DNP3MasterAddr     int    `json:"dnp3_master_addr"`     // this master's link addr (typical 1)

	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`

	// Runtime (not persisted)
	State   LineState `json:"state"`
	RxCount int64     `json:"rx_count"`
	TxCount int64     `json:"tx_count"`
	Addr    string    `json:"addr"`
}

type SignalKind string

const (
	SignalKindDigital SignalKind = "digital"
	SignalKindAnalog  SignalKind = "analog"
)

type Signal struct {
	ID          int64      `json:"id"`
	LineID      int64      `json:"line_id"`
	Name        string     `json:"name"`
	IOA         int        `json:"ioa"`
	TypeID      int        `json:"type_id"`
	Kind        SignalKind `json:"signal_type"`
	Unit        string     `json:"unit"`
	Scale       float64    `json:"scale"`
	Offset      float64    `json:"offset"`
	Description string     `json:"description"`

	// PointType is the DNP3 object group for "dnp3" lines (binary|analog|counter|
	// double_bit_binary|binary_output_status|analog_output_status|...); empty for
	// IEC-104 lines. For DNP3, IOA is reused as the point index within its type.
	PointType string `json:"point_type"`

	LineName string `json:"line_name,omitempty"` // populated by ListAllSignals
}

// IsDNP3 reports whether this line uses the DNP3 driver.
func (l Line) IsDNP3() bool { return l.Protocol == ProtocolDNP3 }

type Datapoint struct {
	SignalID   int64      `json:"signal_id"`
	LineID     int64      `json:"line_id"`
	IOA        int        `json:"ioa"`
	Name       string     `json:"name"`
	Kind       SignalKind `json:"signal_type"`
	TypeID     int        `json:"type_id"`
	Unit       string     `json:"unit"`
	Value      float64    `json:"value"`
	RawValue   float64    `json:"raw_value"`
	Quality    uint8      `json:"quality"`
	Timestamp  time.Time  `json:"timestamp"`
	ReceivedAt time.Time  `json:"received_at"`
}

// QualityOK returns true when IV and BL bits are clear.
func (d *Datapoint) QualityOK() bool {
	return d.Quality&0x80 == 0 && d.Quality&0x10 == 0
}

// Signal includes an optional LineName populated by ListAllSignals.
// The field is omitted from normal per-line queries.

// ScadaView is a named P&ID diagram stored as a JSON layout.
type ScadaView struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Width     int             `json:"width"`
	Height    int             `json:"height"`
	Elements  json.RawMessage `json:"elements"` // []ScadaElement JSON, stored as TEXT in SQLite
	Lines     json.RawMessage `json:"lines"`    // []ScadaLine JSON, stored as TEXT in SQLite
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type Command struct {
	LineID int64   `json:"line_id"`
	IOA    int     `json:"ioa"`
	TypeID int     `json:"type_id"`
	Value  float64 `json:"value"`
	Select bool    `json:"select"`
}

// HistoryPoint is one time-series sample stored in signal_history.
type HistoryPoint struct {
	TS      float64 `json:"ts"` // Unix epoch seconds (float64 for sub-second precision)
	Value   float64 `json:"value"`
	Quality uint8   `json:"quality"`
}
