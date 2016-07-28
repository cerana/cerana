package acomm

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
	logx "github.com/cerana/cerana/pkg/logrusx"
)

// NewStreamUnix sets up an ad-hoc unix listner to stream data.
func (t *Tracker) NewStreamUnix(dir string, src io.ReadCloser) (*url.URL, error) {
	if src == nil {
		err := errors.New("missing stream src")
		log.WithFields(log.Fields{
			"error": err,
		}).Error(err)
		return nil, err
	}

	socketPath, err := generateTempSocketPath(dir, "")
	if err != nil {
		return nil, err
	}

	ul := NewUnixListener(socketPath, 1)
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
func (t *Tracker) ProxyStreamHTTPURL(addr *url.URL) (*url.URL, error) {
	if t.httpStreamURL == nil {
		err := errors.New("tracker missing http stream url")
		log.WithFields(log.Fields{
			"error": err,
		}).Error(err)
		return nil, err
	}

	if addr == nil {
		err := errors.New("missing addr")
		log.WithFields(log.Fields{
			"error": err,
		}).Error(err)
		return nil, err
	}
	streamAddr := &url.URL{}
	*streamAddr = *t.httpStreamURL
	q := streamAddr.Query()
	q.Set("addr", addr.String())
	streamAddr.RawQuery = q.Encode()

	return streamAddr, nil
}

// Stream streams data from a URL to a destination writer.
func Stream(dest io.Writer, addr *url.URL) error {
	if dest == nil {
		err := errors.New("missing dest")
		log.WithFields(log.Fields{
			"error": err,
		}).Error(err)
		return err
	}
	if addr == nil {
		err := errors.New("missing addr")
		log.WithFields(log.Fields{
			"error": err,
		}).Error(err)
		return err
	}

	switch addr.Scheme {
	case "unix":
		return streamUnix(dest, addr)
	case "http", "https":
		return streamHTTP(dest, addr)
	default:
		err := errors.New("unknown url type")
		log.WithFields(log.Fields{
			"error": err,
			"type":  addr.Scheme,
			"addr":  addr,
		}).Error("cannot stream from url")
		return err
	}
}

// streamUnix streams data from a unix socket to a destination writer.
func streamUnix(dest io.Writer, addr *url.URL) error {
	conn, err := net.Dial("unix", addr.RequestURI())
	if err != nil {
		log.WithFields(log.Fields{
			"addr":  addr,
			"error": err,
		}).Error("failed to connect to stream socket")
		return err
	}
	defer logx.LogReturnedErr(conn.Close,
		log.Fields{"addr": addr},
		"failed to close stream connection",
	)

	if _, err := io.Copy(dest, conn); err != nil {
		log.WithFields(log.Fields{
			"addr":  addr,
			"error": err,
		}).Error("failed to stream data")
		return err
	}
	return nil
}

// streamHTTP streams data from an http connection to a destination writer.
func streamHTTP(dest io.Writer, addr *url.URL) error {
	httpResp, err := http.Get(addr.String())
	if err != nil {
		log.WithFields(log.Fields{
			"addr":  addr,
			"error": err,
		}).Error("failed to GET stream")
		return err
	}
	defer logx.LogReturnedErr(httpResp.Body.Close,
		log.Fields{"addr": addr},
		"failed to close stream response body",
	)

	if _, err := io.Copy(dest, httpResp.Body); err != nil {
		log.WithFields(log.Fields{
			"addr":  addr,
			"error": err,
		}).Error("failed to stream data")
		return err
	}
	return nil
}

// ProxyStreamHandler is an HTTP HandlerFunc for simple proxy streaming.
func ProxyStreamHandler(w http.ResponseWriter, r *http.Request) {
	log.WithField("addr", r.URL.Query().Get("addr")).Debug("proxy stream handler addr")

	addr, err := url.ParseRequestURI(r.URL.Query().Get("addr"))
	if err != nil {
		http.Error(w, "invalid addr", http.StatusBadRequest)
		return
	}

	log.WithField("addr", addr).Info("proxying stream")

	if err := Stream(w, addr); err != nil {
		if opErr, ok := err.(*net.OpError); ok {
			// TODO: find out what the result is for not exist and return 404
			log.WithFields(log.Fields{
				"addr":  addr,
				"error": opErr,
			}).Error("failed to proxy stream")
		} else {
			log.WithFields(log.Fields{
				"addr":  addr,
				"error": err,
			}).Error("failed to proxy stream")
		}
		http.Error(w, "failed to stream data", http.StatusInternalServerError)
		return
	}
}
