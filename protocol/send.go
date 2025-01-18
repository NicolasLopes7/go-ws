package protocol

import (
	"encoding/binary"
	"math"
)

// Send takes a `Frame`, builds the appropriate frame format based on
// RFC 6455 specifications, and sends it to the client or server.
//
// The frame format includes the following components:
// - FIN and Opcode: Indicate the type and finality of the frame.
// - Payload Length: Defines the size of the payload, including extended lengths if necessary.
// - Payload Data: Contains the actual application data.
func (ws *Ws) Send(frame Frame) error {
	// Initialize the frame header with 2 bytes.
	// Byte 1: FIN (1 bit) | Reserved (3 bits) | Opcode (4 bits)
	data := make([]byte, 2)
	data[0] = 0x80 | frame.Opcode
	if frame.IsFragment {
		data[0] &= 0x7f
	}

	// Byte 2: Payload length and additional length data if required.
	if frame.Length <= 125 {
		// If payload length fits in 7 bits, encode it directly.
		data[1] = byte(frame.Length)
		data = append(data, frame.Payload...)

		// For payload lengths between 126 and 2^16-1, use 2 bytes for extended length.
	} else if frame.Length > 125 && float64(frame.Length) < math.Pow(2, 16) {
		data[1] = byte(126)
		size := make([]byte, 2)
		binary.BigEndian.PutUint16(size, uint16(frame.Length))
		data = append(data, size...)
		data = append(data, frame.Payload...)
		// For payload lengths >= 2^16, use 8 bytes for extended length.
	} else if float64(frame.Length) >= math.Pow(2, 16) {
		data[1] = byte(127)
		size := make([]byte, 8)
		binary.BigEndian.PutUint64(size, frame.Length)
		data = append(data, size...)
		data = append(data, frame.Payload...)
	}

	return ws.write(data)
}
