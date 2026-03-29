package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
)

// OSCClient sends OSC messages to TotalMix FX
type OSCClient struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

func NewOSCClient(host string, port int) (*OSCClient, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	return &OSCClient{conn: conn, addr: addr}, nil
}

// SendFloat sends an OSC message with a single float32 argument
func (c *OSCClient) SendFloat(address string, value float32) error {
	msg := buildOSCMessage(address, value)
	_, err := c.conn.Write(msg)
	return err
}

func (c *OSCClient) Close() {
	c.conn.Close()
}

// buildOSCMessage constructs a minimal OSC message with one float32 argument.
// OSC format: address (null-padded to 4-byte boundary), type tag ",f\0\0", float32 big-endian
func buildOSCMessage(address string, value float32) []byte {
	var buf []byte

	// Address string (null-terminated, padded to 4 bytes)
	buf = append(buf, []byte(address)...)
	buf = append(buf, 0)
	for len(buf)%4 != 0 {
		buf = append(buf, 0)
	}

	// Type tag string ",f" (null-terminated, padded to 4 bytes)
	buf = append(buf, ',', 'f', 0, 0)

	// Float32 argument (big-endian)
	var fb [4]byte
	binary.BigEndian.PutUint32(fb[:], math.Float32bits(value))
	buf = append(buf, fb[:]...)

	return buf
}

// OSCListener listens for OSC messages from TotalMix FX to sync state
type OSCListener struct {
	conn    *net.UDPConn
	volume  *VolumeState
}

func NewOSCListener(port int, volume *VolumeState) (*OSCListener, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	return &OSCListener{conn: conn, volume: volume}, nil
}

// Listen reads incoming OSC messages and updates volume state.
// This blocks forever.
func (l *OSCListener) Listen() {
	buf := make([]byte, 4096)
	for {
		n, _, err := l.conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		l.handleMessage(buf[:n])
	}
}

func (l *OSCListener) handleMessage(data []byte) {
	// Check if this is an OSC bundle
	if len(data) >= 8 && string(data[:8]) == "#bundle\x00" {
		l.handleBundle(data)
		return
	}
	l.handleSingleMessage(data)
}

func (l *OSCListener) handleBundle(data []byte) {
	// OSC bundle format:
	//   "#bundle\0" (8 bytes)
	//   timetag (8 bytes)
	//   then repeated: size (4 bytes big-endian) + element (size bytes)
	pos := 16 // skip "#bundle\0" + timetag
	for pos+4 <= len(data) {
		size := int(binary.BigEndian.Uint32(data[pos : pos+4]))
		pos += 4
		if pos+size > len(data) || size <= 0 {
			break
		}
		// Each element can be a message or nested bundle
		l.handleMessage(data[pos : pos+size])
		pos += size
	}
}

func (l *OSCListener) handleSingleMessage(data []byte) {
	address, value, ok := parseOSCFloat(data)
	if !ok {
		return
	}
	if address == "/1/mastervolume" {
		l.volume.SetLevel(value)
	}
}

func (l *OSCListener) Close() {
	l.conn.Close()
}

// parseOSCFloat parses a minimal OSC message with one float32 argument
func parseOSCFloat(data []byte) (address string, value float32, ok bool) {
	if len(data) < 8 {
		return "", 0, false
	}

	// Read address (null-terminated)
	addrEnd := 0
	for addrEnd < len(data) && data[addrEnd] != 0 {
		addrEnd++
	}
	address = string(data[:addrEnd])

	// Skip to next 4-byte boundary
	pos := addrEnd + 1
	for pos%4 != 0 {
		pos++
	}

	// Check type tag
	if pos >= len(data) || data[pos] != ',' {
		return "", 0, false
	}
	// Find the float type
	tagEnd := pos
	for tagEnd < len(data) && data[tagEnd] != 0 {
		tagEnd++
	}
	tagStr := string(data[pos:tagEnd])
	if tagStr != ",f" {
		return "", 0, false
	}

	// Skip type tag to next 4-byte boundary
	pos = tagEnd + 1
	for pos%4 != 0 {
		pos++
	}

	// Read float32
	if pos+4 > len(data) {
		return "", 0, false
	}
	bits := binary.BigEndian.Uint32(data[pos : pos+4])
	value = math.Float32frombits(bits)

	return address, value, true
}
