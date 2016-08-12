package acomm

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"

	"github.com/cerana/cerana/pkg/errors"
)

// UnmarshalConnData reads and unmarshals JSON data from the connection into
// the destination object.
func UnmarshalConnData(conn net.Conn, dest interface{}) error {
	sizeBytes := make([]byte, 4)
	if _, err := conn.Read(sizeBytes); err != nil {
		return errors.Wrap(err)
	}
	payloadSize := binary.BigEndian.Uint32(sizeBytes)

	// wrap in an io.LimitReader with the expected size to prevent
	// malformed payloads from blocking the decoder
	payloadReader := io.LimitReader(conn, int64(payloadSize))

	decoder := json.NewDecoder(payloadReader)

	return errors.Wrap(decoder.Decode(dest))
}

// SendConnData marshals and writes payload JSON data to the Conn with
// appropriate headers.
func SendConnData(conn net.Conn, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"payload": payload})
	}

	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, uint32(len(payloadJSON)))

	data := append(sizeBytes, payloadJSON...)

	if _, err := conn.Write(data); err != nil {
		return errors.Wrapv(err, map[string]interface{}{
			"addr": conn.RemoteAddr(),
			"data": string(data),
		})
	}

	return nil
}
