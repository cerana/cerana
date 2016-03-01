package acomm

import (
	"encoding/binary"
	"encoding/json"
	"net"

	log "github.com/Sirupsen/logrus"
)

// UnmarshalConnData reads and unmarshals JSON data from the connection into
// the destination object.
func UnmarshalConnData(conn net.Conn, dest interface{}) error {
	sizeBytes := make([]byte, 4)
	if _, err := conn.Read(sizeBytes); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("failed to read size header")
		return err
	}
	payloadSize := binary.BigEndian.Uint32(sizeBytes)

	data := make([]byte, int(payloadSize))
	if _, err := conn.Read(data); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("failed to read data")
		return err
	}

	if err := json.Unmarshal(data, dest); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"data":  string(data),
		}).Error("failed to unmarshal data")
		return err
	}

	return nil
}

// SendConnData marshals and writes payload JSON data to the Conn with
// appropriate headers.
func SendConnData(conn net.Conn, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"payload": payload,
		}).Error("failed to marshal payload json")
		return err
	}

	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, uint32(len(payloadJSON)))

	data := append(sizeBytes, payloadJSON...)

	if _, err := conn.Write(data); err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"addr":    conn.RemoteAddr(),
			"payload": payload,
		}).Error("failed to write payload to unix socket")
		return err
	}

	return nil
}
