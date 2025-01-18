package protocol

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"
)

// getAcceptHashKey calculates the `Sec-WebSocket-Accept` from the `Sec-WebSocket-Key`
// You can find at https://datatracker.ietf.org/doc/html/rfc6455#section-1.3 how the header's calculated.
// For this header field, the server has to take the value (as present
// in the header field, e.g., the base64-encoded [RFC4648] version minus
// any leading and trailing whitespace) and concatenate this with the
// Globally Unique Identifier (GUID, [RFC4122]) "258EAFA5-E914-47DA-
// 95CA-C5AB0DC85B11" in string form, which is unlikely to be used by
// network endpoints that do not understand the WebSocket Protocol.  A
// SHA-1 hash (160 bits) [FIPS.180-3], base64-encoded (see Section 4 of
// [RFC4648]), of this concatenation is then returned in the server's
// handshake.
func getAcceptHash(key string) string {
	h := sha1.New()
	h.Write([]byte(key))
	h.Write([]byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// To complete the handshake, the server must send a response back
// with the appropriate headers, the handshake will look like this:
// `Sec-WebSocket-Accept` is a special header that's calculated on `getAcceptHashKey`
// HTTP/1.1 101 Switching Protocols
// Upgrade: websocket
// Connection: Upgrade
// Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
func (ws *Ws) Handshake() error {
	hash := getAcceptHash(ws.header.Get("Sec-WebSocket-Key"))

	lines := []string{
		"HTTP/1.1 101 Switching Protocols",
		"Upgrade: websocket",
		"Connection: Upgrade",
		"Server: youdid/it",
		fmt.Sprintf("Sec-WebSocket-Accept: %s", hash),
		"",
		"",
	}

	return ws.write([]byte(strings.Join(lines, "\r\n")))
}
