package iec104

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const startByte = 0x68

type FrameType uint8

const (
	FrameTypeI FrameType = iota
	FrameTypeS
	FrameTypeU
)

type UFunc uint8

const (
	STARTDT_ACT UFunc = 0x07
	STARTDT_CON UFunc = 0x0B
	STOPDT_ACT  UFunc = 0x13
	STOPDT_CON  UFunc = 0x23
	TESTFR_ACT  UFunc = 0x43
	TESTFR_CON  UFunc = 0x83
)

// APCI holds the decoded Application Protocol Control Information.
type APCI struct {
	Type  FrameType
	SSN   uint16 // I-frame send seq num (15-bit)
	RSN   uint16 // I/S-frame recv seq num (15-bit)
	UFunc UFunc
	ASDU  []byte // present only in I-frames
}

// ReadFrame reads one complete APDU from conn and decodes its control field.
func ReadFrame(conn net.Conn) (*APCI, error) {
	hdr := make([]byte, 2)
	if _, err := io.ReadFull(conn, hdr); err != nil {
		return nil, err
	}
	if hdr[0] != startByte {
		return nil, fmt.Errorf("invalid start byte 0x%02X", hdr[0])
	}
	length := int(hdr[1])
	if length < 4 {
		return nil, fmt.Errorf("APDU length %d < 4", length)
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}
	return parseControl(buf)
}

func parseControl(buf []byte) (*APCI, error) {
	if len(buf) < 4 {
		return nil, fmt.Errorf("control field too short")
	}
	a := &APCI{}
	cf1 := buf[0]

	switch {
	case cf1&0x01 == 0: // I-frame: bit0 = 0
		a.Type = FrameTypeI
		a.SSN = (uint16(buf[0])>>1 | uint16(buf[1])<<7) & 0x7FFF
		a.RSN = (uint16(buf[2])>>1 | uint16(buf[3])<<7) & 0x7FFF
		if len(buf) > 4 {
			a.ASDU = buf[4:]
		}
	case cf1&0x03 == 0x01: // S-frame: bits1-0 = 01
		a.Type = FrameTypeS
		a.RSN = (uint16(buf[2])>>1 | uint16(buf[3])<<7) & 0x7FFF
	default: // U-frame: bits1-0 = 11
		a.Type = FrameTypeU
		a.UFunc = UFunc(cf1)
	}
	return a, nil
}

// BuildIFrame encodes an I-frame APDU.
func BuildIFrame(ssn, rsn uint16, asdu []byte) []byte {
	length := 4 + len(asdu)
	buf := make([]byte, 2+length)
	buf[0] = startByte
	buf[1] = byte(length)
	binary.LittleEndian.PutUint16(buf[2:4], (ssn&0x7FFF)<<1)
	binary.LittleEndian.PutUint16(buf[4:6], (rsn&0x7FFF)<<1)
	copy(buf[6:], asdu)
	return buf
}

// BuildSFrame encodes a supervisory frame.
func BuildSFrame(rsn uint16) []byte {
	buf := make([]byte, 6)
	buf[0] = startByte
	buf[1] = 4
	buf[2] = 0x01
	buf[3] = 0x00
	binary.LittleEndian.PutUint16(buf[4:6], (rsn&0x7FFF)<<1)
	return buf
}

// BuildUFrame encodes an unnumbered frame.
func BuildUFrame(fn UFunc) []byte {
	buf := make([]byte, 6)
	buf[0] = startByte
	buf[1] = 4
	buf[2] = byte(fn)
	return buf
}

// SeqDiff returns the forward distance from b to a in 15-bit space.
func SeqDiff(a, b uint16) uint16 {
	return (a - b + 32768) % 32768
}
