package protocol

import "encoding/binary"

// Recv reads and decodes a single WebSocket frame from the connection.
// The frame format is defined in the WebSocket protocol (RFC 6455).
// Frame Specification: https://datatracker.ietf.org/doc/html/rfc6455#section-5.2
// Masking Specification: https://datatracker.ietf.org/doc/html/rfc6455#section-5.3
//
// A WebSocket frame consists of the following components:
//   - 1st Byte: Metadata flags
//   - FIN (1 bit): Indicates if this is the final fragment of the message.
//   - RSV1/RSV2/RSV3 (1 bit each): Reserved bits for extensions.
//   - Opcode (4 bits): Defines the frame type (e.g., text, binary, close).
//   - 2nd Byte: Payload and Mask Metadata
//   - MASK (1 bit): Indicates if the payload is masked (always true for client frames).
//   - Payload Length (7 bits or extended length fields).
//   - Extended Payload Length: (if applicable, 2 or 8 bytes for larger payloads).
//   - Masking Key (4 bytes, if MASK is set): Used to decode the client payload.
//   - Payload Data: The actual application data, which is masked or unmasked.
func (ws *Ws) Recv() (Frame, error) {
	frame := Frame{}

	// Read the first 2 bytes of the frame header.
	head, err := ws.read(2)
	if err != nil {
		return frame, err
	}

	// Decode the frame metadata from the first byte:
	// FIN
	frame.IsFragment = (head[0] & 0x80) == 0x00
	// 4-bit opcode
	frame.Opcode = head[0] & 0x0F
	// reserved bits (RSV1/RSV2/RSV3)
	frame.Reserved = (head[0] & 0x70)

	// Decode the frame metadata from the second byte:
	frame.IsMasked = (head[1] & 0x80) == 0x80 // Check if the payload is masked (MASK = 1).
	length := uint64(head[1] & 0x7F)          // Extract the initial 7-bit payload length.

	switch length {
	case 126: // Payload length is in the next 2 bytes (16-bit unsigned).
		data, err := ws.read(2)
		if err != nil {
			return frame, err
		}
		length = uint64(binary.BigEndian.Uint16(data))
	case 127: // Payload length is in the next 8 bytes (64-bit unsigned).
		data, err := ws.read(8)
		if err != nil {
			return frame, err
		}
		length = binary.BigEndian.Uint64(data)
	}
	frame.Length = length

	// Read the 4-byte masking key (if MASK is set).
	var mask []byte
	if frame.IsMasked {
		mask, err = ws.read(4)
		if err != nil {
			return frame, err
		}
	}

	payload, err := ws.read(int(length))
	if err != nil {
		return frame, err
	}

	if frame.IsMasked {
		// Decode the payload using the masking key (repeats every 4 bytes).
		for i := uint64(0); i < length; i++ {
			payload[i] ^= mask[i%4]
		}
	}

	frame.Payload = payload
	return frame, nil
}
