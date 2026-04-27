package iec104

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"go104/internal/models"
)

// ASDU holds a decoded Application Service Data Unit.
type ASDU struct {
	TypeID  uint8
	SQ      bool   // sequence flag
	Count   int    // number of info objects
	T       bool   // test flag
	PN      bool   // positive/negative
	COT     uint8  // cause of transmission
	OA      uint8  // originator address
	CA      uint16 // common address
	Objects []InfoObject
}

// InfoObject holds one decoded information object.
type InfoObject struct {
	IOA       uint32
	Value     float64
	Quality   uint8
	Timestamp time.Time
	HasTime   bool
}

// ParseASDU decodes raw ASDU bytes from an I-frame payload.
func ParseASDU(data []byte) (*ASDU, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("ASDU too short: %d bytes", len(data))
	}
	a := &ASDU{
		TypeID: data[0],
		SQ:     data[1]&0x80 != 0,
		Count:  int(data[1] & 0x7F),
		T:      data[2]&0x80 != 0,
		PN:     data[2]&0x40 != 0,
		COT:    data[2] & 0x3F,
		OA:     data[3],
		CA:     binary.LittleEndian.Uint16(data[4:6]),
	}

	objs, err := parseObjects(a.TypeID, a.SQ, a.Count, data[6:])
	if err != nil {
		return nil, fmt.Errorf("type %d: %w", a.TypeID, err)
	}
	a.Objects = objs
	return a, nil
}

// objectDataSize returns byte size of one object's DATA (after IOA).
func objectDataSize(typeID uint8) int {
	switch int(typeID) {
	case M_SP_NA_1:
		return 1
	case M_DP_NA_1:
		return 1
	case M_ME_NA_1:
		return 3 // NVA(2)+QDS(1)
	case M_ME_NB_1:
		return 3 // SVA(2)+QDS(1)
	case M_ME_NC_1:
		return 5 // IEEE754(4)+QDS(1)
	case M_ME_ND_1:
		return 2 // NVA(2), no QDS
	case M_SP_TB_1:
		return 8 // SIQ(1)+CP56(7)
	case M_DP_TB_1:
		return 8 // DIQ(1)+CP56(7)
	case M_ME_TA_1:
		return 10 // NVA(2)+QDS(1)+CP56(7)
	case M_ME_TB_1:
		return 10 // SVA(2)+QDS(1)+CP56(7)
	case M_ME_TF_1:
		return 12 // IEEE754(4)+QDS(1)+CP56(7)
	case C_IC_NA_1:
		return 1 // QOI
	case C_CS_NA_1:
		return 7 // CP56
	}
	return -1
}

func parseObjects(typeID uint8, sq bool, count int, data []byte) ([]InfoObject, error) {
	dataSize := objectDataSize(typeID)
	if dataSize < 0 {
		return nil, fmt.Errorf("unsupported TypeID %d", typeID)
	}

	objs := make([]InfoObject, 0, count)

	if sq {
		// Sequence: first 3 bytes = base IOA, then count * dataSize
		if len(data) < 3+count*dataSize {
			return nil, fmt.Errorf("SQ frame too short")
		}
		baseIOA := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16
		pos := 3
		for i := 0; i < count; i++ {
			obj := parseObjectData(typeID, data[pos:pos+dataSize])
			obj.IOA = baseIOA + uint32(i)
			objs = append(objs, obj)
			pos += dataSize
		}
	} else {
		// Non-sequence: each object has IOA + data
		stride := 3 + dataSize
		if len(data) < count*stride {
			return nil, fmt.Errorf("frame too short for %d objects", count)
		}
		for i := 0; i < count; i++ {
			pos := i * stride
			ioa := uint32(data[pos]) | uint32(data[pos+1])<<8 | uint32(data[pos+2])<<16
			obj := parseObjectData(typeID, data[pos+3:pos+stride])
			obj.IOA = ioa
			objs = append(objs, obj)
		}
	}
	return objs, nil
}

func parseObjectData(typeID uint8, d []byte) InfoObject {
	var o InfoObject
	switch int(typeID) {
	case M_SP_NA_1:
		o.Quality = d[0] & 0xF0
		o.Value = float64(d[0] & 0x01)
	case M_DP_NA_1:
		o.Quality = d[0] & 0xF0
		o.Value = float64(d[0] & 0x03)
	case M_ME_NA_1:
		o.Value = float64(int16(binary.LittleEndian.Uint16(d[0:2]))) / 32768.0
		o.Quality = d[2]
	case M_ME_NB_1:
		o.Value = float64(int16(binary.LittleEndian.Uint16(d[0:2])))
		o.Quality = d[2]
	case M_ME_NC_1:
		o.Value = float64(math.Float32frombits(binary.LittleEndian.Uint32(d[0:4])))
		o.Quality = d[4]
	case M_ME_ND_1:
		o.Value = float64(int16(binary.LittleEndian.Uint16(d[0:2]))) / 32768.0
	case M_SP_TB_1:
		o.Quality = d[0] & 0xF0
		o.Value = float64(d[0] & 0x01)
		o.Timestamp = parseCP56Time2a(d[1:8])
		o.HasTime = true
	case M_DP_TB_1:
		o.Quality = d[0] & 0xF0
		o.Value = float64(d[0] & 0x03)
		o.Timestamp = parseCP56Time2a(d[1:8])
		o.HasTime = true
	case M_ME_TA_1:
		o.Value = float64(int16(binary.LittleEndian.Uint16(d[0:2]))) / 32768.0
		o.Quality = d[2]
		o.Timestamp = parseCP56Time2a(d[3:10])
		o.HasTime = true
	case M_ME_TB_1:
		o.Value = float64(int16(binary.LittleEndian.Uint16(d[0:2])))
		o.Quality = d[2]
		o.Timestamp = parseCP56Time2a(d[3:10])
		o.HasTime = true
	case M_ME_TF_1:
		o.Value = float64(math.Float32frombits(binary.LittleEndian.Uint32(d[0:4])))
		o.Quality = d[4]
		o.Timestamp = parseCP56Time2a(d[5:12])
		o.HasTime = true
	}
	return o
}

// parseCP56Time2a decodes a 7-byte CP56Time2a timestamp.
func parseCP56Time2a(b []byte) time.Time {
	if len(b) < 7 || b[2]&0x80 != 0 { // invalid bit set
		return time.Now().UTC()
	}
	ms := int(b[0]) | int(b[1])<<8
	min := int(b[2] & 0x3F)
	hour := int(b[3] & 0x1F)
	dom := int(b[4] & 0x1F)
	month := int(b[5] & 0x0F)
	year := 2000 + int(b[6]&0x7F)
	return time.Date(year, time.Month(month), dom, hour, min, ms/1000, (ms%1000)*1_000_000, time.UTC)
}

// encodeCP56Time2a encodes a time.Time into 7-byte CP56Time2a.
func encodeCP56Time2a(t time.Time) []byte {
	t = t.UTC()
	ms := t.Second()*1000 + t.Nanosecond()/1_000_000
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}
	return []byte{
		byte(ms & 0xFF),
		byte(ms >> 8),
		byte(t.Minute()),
		byte(t.Hour()),
		byte(t.Day()) | byte(wd<<5),
		byte(t.Month()),
		byte(t.Year() % 100),
	}
}

// IsMonitoringType returns true for data-carrying TypeIDs.
func IsMonitoringType(typeID int) bool {
	return IsDigitalType(typeID) || IsAnalogType(typeID)
}

// --- ASDU builders ---

// BuildGIAsdu builds a C_IC_NA_1 (general interrogation) ASDU.
func BuildGIAsdu(ca uint16) []byte {
	b := make([]byte, 10)
	b[0] = C_IC_NA_1
	b[1] = 0x01
	b[2] = CotActivation
	b[3] = 0x00
	binary.LittleEndian.PutUint16(b[4:6], ca)
	// IOA = 0
	b[9] = 20 // QOI: station interrogation
	return b
}

// BuildClockSyncAsdu builds a C_CS_NA_1 ASDU with current time.
func BuildClockSyncAsdu(ca uint16, t time.Time) []byte {
	b := make([]byte, 16)
	b[0] = C_CS_NA_1
	b[1] = 0x01
	b[2] = CotActivation
	b[3] = 0x00
	binary.LittleEndian.PutUint16(b[4:6], ca)
	// IOA = 0 (bytes 6-8 = 0)
	copy(b[9:16], encodeCP56Time2a(t))
	return b
}

// BuildCommandAsdu builds a command ASDU for the given Command.
func BuildCommandAsdu(cmd models.Command, ca uint16) []byte {
	ioa := cmd.IOA
	ioab := [3]byte{byte(ioa), byte(ioa >> 8), byte(ioa >> 16)}

	hdr := [6]byte{}
	hdr[1] = 0x01 // count=1
	hdr[2] = CotActivation
	binary.LittleEndian.PutUint16(hdr[4:6], ca)

	switch cmd.TypeID {
	case C_SC_NA_1:
		hdr[0] = C_SC_NA_1
		sco := byte(0)
		if cmd.Value != 0 {
			sco |= 0x01
		}
		if cmd.Select {
			sco |= 0x80
		}
		return append(append(hdr[:], ioab[:]...), sco)

	case C_DC_NA_1:
		hdr[0] = C_DC_NA_1
		dco := byte(cmd.Value) & 0x03
		if cmd.Select {
			dco |= 0x80
		}
		return append(append(hdr[:], ioab[:]...), dco)

	case C_SE_NA_1:
		hdr[0] = C_SE_NA_1
		v := int16(cmd.Value * 32768)
		qos := byte(0)
		if cmd.Select {
			qos |= 0x80
		}
		b := []byte{byte(v), byte(uint16(v) >> 8), qos}
		return append(append(hdr[:], ioab[:]...), b...)

	case C_SE_NB_1:
		hdr[0] = C_SE_NB_1
		v := int16(cmd.Value)
		qos := byte(0)
		if cmd.Select {
			qos |= 0x80
		}
		b := []byte{byte(v), byte(uint16(v) >> 8), qos}
		return append(append(hdr[:], ioab[:]...), b...)

	case C_SE_NC_1:
		hdr[0] = C_SE_NC_1
		fb := math.Float32bits(float32(cmd.Value))
		qos := byte(0)
		if cmd.Select {
			qos |= 0x80
		}
		b := make([]byte, 5)
		binary.LittleEndian.PutUint32(b[0:4], fb)
		b[4] = qos
		return append(append(hdr[:], ioab[:]...), b...)
	}
	return nil
}
