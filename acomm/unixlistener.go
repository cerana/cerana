package acomm

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/pkg/logrusx"
)

// UnixListener is a wrapper for a unix socket. It handles creation and
// listening for new connections, as well as graceful shutdown.
type UnixListener struct {
	acceptLimit int
	addr        *net.UnixAddr
	listener    *net.UnixListener
	waitgroup   sync.WaitGroup
	stopChan    chan struct{}
	connChan    chan net.Conn
}

// NewUnixListener creates and initializes a new UnixListener. AcceptLimit
// controls how many connections it will listen for before stopping; 0 and
// below is unlimited.
func NewUnixListener(socketPath string, acceptLimit int) *UnixListener {
	// Ignore error since the only time it would arise is with a bad net
	// parameter
	addr, _ := net.ResolveUnixAddr("unix", socketPath)

	// Negatives are easier to work with for unlimited than zero
	if acceptLimit <= 0 {
		acceptLimit = -1
	}

	return &UnixListener{
		addr:     addr,
		stopChan: make(chan struct{}),
		// Note: The chan here just holds conns until they get passed to a
		// handler. The buffer size does not control conn handling concurrency.
		connChan:    make(chan net.Conn, 1000),
		acceptLimit: acceptLimit,
	}
}

// Addr returns the string representation of the unix address.
func (ul *UnixListener) Addr() string {
	return ul.addr.String()
}

// URL returns the URL representation of the unix address.
func (ul *UnixListener) URL() *url.URL {
	u, _ := url.ParseRequestURI(fmt.Sprintf("unix://%s", ul.Addr()))
	return u
}

// Start prepares the listener and starts listening for new connections.
func (ul *UnixListener) Start() error {
	if err := ul.createListener(); err != nil {
		return err
	}

	ul.stopChan = make(chan struct{})

	// Waitgroup should wait for the listener itself to close
	ul.waitgroup.Add(1)
	go ul.listen()

	return nil
}

// createListener creates a new net.UnixListener
func (ul *UnixListener) createListener() error {
	// create directory structure if it does not exist yet
	directory := filepath.Dir(ul.Addr())
	// TODO: Decide on permissions
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		return errors.Wrapv(err, map[string]interface{}{
			"directory": directory,
			"perm":      os.ModePerm,
		})
	}

	listener, err := net.ListenUnix("unix", ul.addr)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"addr": ul.Addr()})
	}

	ul.listener = listener
	return nil
}

// listen continuously listens and accepts new connections up to the accept
// limit.
func (ul *UnixListener) listen() {
	defer ul.waitgroup.Done()
	defer logrusx.LogReturnedErr(ul.listener.Close, map[string]interface{}{
		"addr": ul.Addr(),
	}, "failed to close listener")

	for i := ul.acceptLimit; i != 0; {
		select {
		case <-ul.stopChan:
			logrus.WithFields(logrus.Fields{
				"addr": ul.Addr(),
			}).Info("stop listening")
			return
		default:
		}

		if err := ul.listener.SetDeadline(time.Now().Add(time.Second)); err != nil {
			err = errors.Wrapv(err, map[string]interface{}{"addr": ul.Addr()})
			logrus.WithField("error", err).Error("failed to set listener deadline")
		}

		conn, err := ul.listener.Accept()
		if nil != err {
			err = errors.Wrapv(err, map[string]interface{}{"addr": ul.Addr()})
			// Don't worry about a timeout
			if opErr, ok := errors.Cause(err).(*net.OpError); ok && opErr.Timeout() {
				continue
			}

			logrus.WithField("error", err).Error("failed to accept new connection")
			continue
		}

		ul.waitgroup.Add(1)
		ul.connChan <- conn

		// Only decrement i when there's a limit it is counting down
		if i > 0 {
			i--
		}
	}
}

// Stop stops listening for new connections. It blocks until existing
// connections are handled and the listener closed.
func (ul *UnixListener) Stop(timeout time.Duration) {
	close(ul.stopChan)
	ul.waitgroup.Wait()
	return
}

// NextConn blocks and returns the next connection. It will return nil when the
// listener is stopped and all existing connections have been handled.
// Connections should be handled in a go routine to take advantage of
// concurrency. When done, the connection MUST be finished with a call to
// DoneConn.
func (ul *UnixListener) NextConn() net.Conn {
	select {
	case <-ul.stopChan:
		return nil
	case conn := <-ul.connChan:
		return conn
	}
}

// DoneConn completes the handling of a connection.
func (ul *UnixListener) DoneConn(conn net.Conn) {
	if conn == nil {
		return
	}
	logrusx.LogReturnedErr(conn.Close,
		map[string]interface{}{
			"addr": ul.addr,
		}, "failed to close unix connection",
	)

	ul.waitgroup.Done()
}
