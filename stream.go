package acomm

import (
	"io"
	"net/url"
	"time"

	log "github.com/Sirupsen/logrus"
)

type stopper interface {
	Stop(time.Duration)
}

// NewStreamUnix sets up an ad-hoc unix listner to stream data.
func (t *Tracker) NewStreamUnix(id string, src io.ReadCloser) (*url.URL, error) {
	socketPath, err := generateTempSocketPath()
	if err != nil {
		return nil, err
	}

	if id == "" {
		id = socketPath
	}

	ul := NewUnixListener(socketPath)
	if err := ul.Start(); err != nil {
		return nil, err
	}

	t.dsLock.Lock()
	t.dataStreams[id] = ul
	t.dsLock.Unlock()

	go func() {
		defer func() {
			_ = src.Close()

			t.dsLock.Lock()
			delete(t.dataStreams, id)
			t.dsLock.Unlock()
		}()

		conn := ul.NextConn()
		if conn == nil {
			return
		}
		defer ul.Stop(time.Millisecond)
		defer ul.DoneConn(conn)

		if _, err := io.Copy(conn, src); err != nil {
			log.WithFields(log.Fields{
				"id":    id,
				"addr":  socketPath,
				"error": err,
			}).Error("failed to stream data")
			return
		}
	}()

	return ul.URL(), nil
}
