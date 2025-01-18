package protocol

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"
)

type Frame struct {
	IsFragment bool
	Opcode     byte
	Reserved   byte
	IsMasked   bool
	Length     uint64
	Payload    []byte
}

func (f Frame) String() string {
	var payloadDisplay string
	if f.Opcode == 0x1 && utf8.Valid(f.Payload) {
		payloadDisplay = fmt.Sprintf(`"%s"`, string(f.Payload))
	} else {
		maxPayloadDisplay := 32
		truncatedPayload := f.Payload
		if len(f.Payload) > maxPayloadDisplay {
			truncatedPayload = f.Payload[:maxPayloadDisplay]
		}
		payloadDisplay = hex.EncodeToString(truncatedPayload)
		if len(f.Payload) > maxPayloadDisplay {
			payloadDisplay += "..."
		}
	}

	return fmt.Sprintf(strings.TrimSpace(`
Frame:
  IsFragment: %t
  Opcode:     0x%02x (%s)
  Reserved:   0x%02x
  IsMasked:   %t
  Length:     %d
  Payload:    %s
`),
		f.IsFragment,
		f.Opcode, opcodeDescription(f.Opcode),
		f.Reserved,
		f.IsMasked,
		f.Length,
		payloadDisplay,
	)
}

func opcodeDescription(opcode byte) string {
	switch opcode {
	case 0x0:
		return "Continuation Frame"
	case 0x1:
		return "Text Frame"
	case 0x2:
		return "Binary Frame"
	case 0x8:
		return "Close Frame"
	case 0x9:
		return "Ping Frame"
	case 0xA:
		return "Pong Frame"
	default:
		return "Reserved/Unknown Frame"
	}
}
