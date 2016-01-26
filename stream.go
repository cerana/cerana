package acomm

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	log "github.com/Sirupsen/logrus"
	logx "github.com/mistifyio/mistify-logrus-ext"
)

type stopper interface {
	Stop(time.Duration)
}

// NewStreamUnix sets up an ad-hoc unix listner to stream data.
func (t *Tracker) NewStreamUnix(src io.ReadCloser) (*url.URL, error) {
	socketPath, err := generateTempSocketPath()
	if err != nil {
		return nil, err
	}

	ul := NewUnixListener(socketPath)
	if err := ul.Start(); err != nil {
		return nil, err
	}

	t.dsLock.Lock()
	t.dataStreams[socketPath] = ul
	t.dsLock.Unlock()

	go func() {
		defer func() {
			_ = src.Close()

			t.dsLock.Lock()
			delete(t.dataStreams, socketPath)
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
				"socketPath": socketPath,
				"error":      err,
			}).Error("failed to stream data")
			return
		}
	}()

	return ul.URL(), nil
}

// ProxyStreamHTTPURL generates the url for proxying streaming data from a unix
// socket.
func (t *Tracker) ProxyStreamHTTPURL(socketPath string) (*url.URL, error) {
	return url.ParseRequestURI(fmt.Sprintf(t.httpStreamURLFormat, socketPath))
}

// ProxyStream streams data from a unix socket to a destination writer, e.g. an
// http.ResponseWriter. If nothing
func (t *Tracker) ProxyStream(dest io.Writer, socketPath string) error {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return err
	}
	defer logx.LogReturnedErr(conn.Close,
		log.Fields{"addr": socketPath},
		"failed to close unix connection",
	)

	if _, err := io.Copy(dest, conn); err != nil {
		log.WithFields(log.Fields{
			"socketPath": socketPath,
			"error":      err,
		}).Error("failed to stream data")
		return err
	}
	return nil
}

// ProxyStreamHandler is an HTTP HandlerFunc for simple proxy streaming.
func (t *Tracker) ProxyStreamHandler(w http.ResponseWriter, r *http.Request) {
	socketPath := r.URL.Query().Get("socket")
	if socketPath == "" {
		http.Error(w, "missing socket", 400)
		return
	}

	_ = t.ProxyStream(w, socketPath)
}
