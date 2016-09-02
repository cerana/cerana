package acomm

import (
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/pkg/logrusx"
)

// NewStreamUnix sets up an ad-hoc unix listner to stream data.
func (t *Tracker) NewStreamUnix(dir string, src io.ReadCloser) (*url.URL, error) {
	if src == nil {
		return nil, errors.New("missing stream src")
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
			logrusx.DebugReturnedErr(src.Close, map[string]interface{}{"socketPath": socketPath}, "failed to close stream source")

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
			err = errors.Wrapv(err, map[string]interface{}{"socketPath": socketPath}, "failed to stream data")
			logrus.WithField("error", err).Debug(err.Error())
			return
		}
	}()

	return ul.URL(), nil
}

// ProxyStreamHTTPURL generates the url for proxying streaming data from a unix
// socket.
func (t *Tracker) ProxyStreamHTTPURL(addr *url.URL) (*url.URL, error) {
	if t.httpStreamURL == nil {
		return nil, errors.New("tracker missing http stream url")
	}

	if addr == nil {
		return nil, errors.New("missing addr")
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
		return errors.New("missing dest")
	}
	if addr == nil {
		return errors.New("missing addr")
	}

	switch addr.Scheme {
	case "unix":
		return streamUnix(dest, addr)
	case "http", "https":
		return streamHTTP(dest, addr)
	default:
		return errors.Newv("unknown url scheme", map[string]interface{}{"addr": addr})
	}
}

// streamUnix streams data from a unix socket to a destination writer.
func streamUnix(dest io.Writer, addr *url.URL) error {
	conn, err := net.Dial("unix", addr.RequestURI())
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"addr": addr})
	}
	defer logrusx.DebugReturnedErr(conn.Close,
		map[string]interface{}{"addr": addr},
		"failed to close stream connection",
	)

	_, err = io.Copy(dest, conn)
	return errors.Wrapv(err, map[string]interface{}{"addr": addr})
}

// streamHTTP streams data from an http connection to a destination writer.
func streamHTTP(dest io.Writer, addr *url.URL) error {
	httpResp, err := http.Get(addr.String())
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"addr": addr})
	}
	defer logrusx.DebugReturnedErr(httpResp.Body.Close,
		map[string]interface{}{"addr": addr},
		"failed to close stream response body",
	)

	_, err = io.Copy(dest, httpResp.Body)
	return errors.Wrapv(err, map[string]interface{}{"addr": addr})
}

// ProxyStreamHandler is an HTTP HandlerFunc for simple proxy streaming.
func ProxyStreamHandler(w http.ResponseWriter, r *http.Request) {
	addr, err := url.ParseRequestURI(r.URL.Query().Get("addr"))
	if err != nil {
		http.Error(w, "invalid addr", http.StatusBadRequest)
		return
	}

	if err := Stream(w, addr); err != nil {
		if _, ok := errors.Cause(err).(*net.OpError); ok {
			// TODO: find out what the result is for "not-exist" and return 404
			logrus.WithField("error", err).Debug("failed to stream data")
		}
		http.Error(w, "failed to stream data", http.StatusInternalServerError)
		return
	}
}
