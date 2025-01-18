package protocol

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
)

// The WebSocket protocol is really simple to implement.
// We basically need to understand HTTP hijacking and binary enc/dec.
// It uses HTTP protocol for the initial handshake, after it
// uses raw TCP to read/write the data. Specification is here:
// https://datatracker.ietf.org/doc/html/rfc6455
// What we'll be doing is basically:
// - Tries to hijack the connection
// - Open the handshake
// - Receive data from client
// - Send data to client
// - Close the handshake
type Ws struct {
	Conn   net.Conn
	buf    *bufio.ReadWriter
	header http.Header
}

// New creates a new WebSocket (`Ws`) instance by upgrading an HTTP connection.
// The function relies on HTTP hijacking to take control of the underlying connection,
// allowing it to bypass the default HTTP handling
func New(w http.ResponseWriter, r *http.Request) (*Ws, error) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("doesn't support HTTP hijacking")
	}

	conn, buf, err := hijacker.Hijack()
	if err != nil {
		return nil, err
	}
	ws := Ws{conn, buf, r.Header}

	err = ws.Handshake()
	if err != nil {
		return nil, err
	}

	return &ws, err
}

func (ws *Ws) write(data []byte) error {
	if _, err := ws.buf.Write(data); err != nil {
		return err
	}
	return ws.buf.Flush()
}

// read reads `n` bytes of data from a WebSocket's buffered reader and returns it as a byte slice.
// It ensures that exactly `n` bytes are read unless an error occurs during the process.
// If the buffered reader (`ws.buf`) has fewer than `n` bytes available, the function will block until the required data is available or an error occurs.
func (ws *Ws) read(n int) ([]byte, error) {
	data := make([]byte, 0)

	// Loop until the required `n` bytes have been read.
	for {
		// Exit the loop if the desired number of bytes has been read.
		if len(data) == n {
			break
		}

		// Calculate the size of the buffer for the current read iteration.
		// Use the smaller of `bufferSize` or the remaining bytes needed to reach `n`.
		size := bufferSize
		remaining := n - len(data)
		if size > remaining {
			size = remaining
		}

		temp := make([]byte, size)
		bytesRead, err := ws.buf.Read(temp)
		if err != nil {
			// EOF is treated as a normal condition, not an error.
			if err != io.EOF {
				return data, err
			}
		}

		data = append(data, temp[:bytesRead]...)
	}

	return data, nil
}

// Close gracefully closes the WebSocket connection by sending a close control frame (Opcode 8)
// as per the WebSocket protocol (RFC 6455) and then closing the underlying network connection.
func (ws *Ws) Close() error {
	f := Frame{
		Opcode:  8,
		Length:  2,
		Payload: make([]byte, 2),
	}

	binary.BigEndian.PutUint16(f.Payload, closeCode)
	if err := ws.Send(f); err != nil {
		return err
	}

	return ws.Conn.Close()
}
