package acomm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	logx "github.com/mistifyio/mistify-logrus-ext"
)

// Tracker keeps track of requests waiting on a response.
type Tracker struct {
	socketDir        string
	responseListener net.Listener
	requestsLock     sync.Mutex // Protects requests
	requests         map[string]*Request
}

// NewTracker creates and initializes a new Tracker. If a socketDir is not
// provided, the response socket will be created in a temporary directory.
func NewTracker(socketDir string) *Tracker {
	return &Tracker{
		socketDir: socketDir,
		requests:  make(map[string]*Request),
	}
}

// NumRequests returns the number of tracked requests
func (t *Tracker) NumRequests() int {
	t.requestsLock.Lock()
	defer t.requestsLock.Unlock()

	return len(t.requests)
}

// StartListener activates the tracker and starts listening for responses.
func (t *Tracker) StartListener() error {
	// Already listening
	if t.responseListener != nil {
		return nil
	}

	if err := t.createListener(); err != nil {
		return err
	}

	go t.listenForResponses()
	return nil
}

// createListener creates a new listener for responses.
func (t *Tracker) createListener() error {
	// Use TempFile to allocate a uniquely named file in either the specified
	// dir or the default temp dir. It is then removed so that the unix socket
	// can be created with that name.
	f, err := ioutil.TempFile(t.socketDir, "acommTracker-")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("failed to create temp file for response socket")
		return err
	}
	_ = f.Close()
	_ = os.Remove(f.Name())

	socketPath := fmt.Sprintf("%s.sock", f.Name())

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.WithFields(log.Fields{
			"error":      err,
			"socketPath": socketPath,
		}).Error("failed to create response listener")
		return err
	}
	t.responseListener = listener

	return nil
}

// listenForResponse continually accepts new responses on the listener.
func (t *Tracker) listenForResponses() {
	for {
		if t.responseListener == nil {
			return
		}
		conn, err := t.responseListener.Accept()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("failed to accept new connection")
			return
		}

		// Handle connection in a goroutine that connections can be served
		// concurrently.
		go t.handleConn(conn)
	}
}

// handleConn handles the response connection and parses the data.
func (t *Tracker) handleConn(conn net.Conn) {
	// Close the response connection
	defer logx.LogReturnedErr(conn.Close, log.Fields{
		"addr": conn.RemoteAddr(),
	}, "failed to close connection")

	data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("failed to read response body")
		return
	}

	resp := &Response{}
	if err := json.Unmarshal(data, resp); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"data":  string(data),
		}).Error("failed to unmarshal response data")
	}

	_ = t.handleResponse(resp)
}

// handleResponse associates a response with a request and forwards the
// response.
func (t *Tracker) handleResponse(resp *Response) error {
	req := t.RetrieveRequest(resp.ID)
	if req == nil {
		err := errors.New("response does not have tracked request")
		log.WithFields(log.Fields{
			"error":    err,
			"response": resp,
		}).Error(err)
		return err
	}

	return req.Respond(resp)
}

// StopListener disallows new requests to be tracked and waits until either all active
// requests are handled or a timeout occurs. The chan returned will be used to
// notify when the Tracker is fully stopped.
func (t *Tracker) StopListener(timeout time.Duration) error {
	// Nothing to do if it's not listening.
	if t.responseListener == nil {
		return nil
	}

	// Close the response connection before returning
	defer func() { t.responseListener = nil }()
	defer logx.LogReturnedErr(t.responseListener.Close, log.Fields{
		"addr": t.responseListener.Addr(),
	}, "failed to close listener")

	// Try to wait for all tracked requests to finish cleanly.
	totalTime := 0 * time.Millisecond
	sleepTime := 10 * time.Millisecond
	for t.NumRequests() > 0 && totalTime < timeout {
		time.Sleep(sleepTime)
		totalTime += sleepTime
	}

	if t.NumRequests() > 0 {
		err := errors.New("timeout")
		log.WithFields(log.Fields{
			"error":     err,
			"remaining": t.NumRequests(),
		}).Error("not all requests finished before stop timeout")
		return err
	}

	return nil
}

// TrackRequest tracks a request. This should only be called directly when the
// Tracker is being used by an original source of requests. Responses should
// then be removed with RetrieveRequest.
func (t *Tracker) TrackRequest(req *Request) {
	t.requestsLock.Lock()
	defer t.requestsLock.Unlock()

	t.requests[req.ID] = req

	return
}

// RetrieveRequest returns a tracked Request based on ID and stops tracking it.
// This should only be called directly when the Tracker is being used by an
// original source of requests.
func (t *Tracker) RetrieveRequest(id string) *Request {
	t.requestsLock.Lock()
	defer t.requestsLock.Unlock()

	if req, ok := t.requests[id]; ok {
		delete(t.requests, id)
		return req
	}

	return nil
}

// ProxyUnix proxies requests that have response hooks of non-unix sockets
// through one that does. If the response hook is already a unix socket, it
// returns the original request. If not, it tracks the original request and
// returns a new request with a unix socket response hook. The purpose of this
// is so that there can be a single entry and exit point for external
// communication, while local services can reply directly to each other.
func (t *Tracker) ProxyUnix(req *Request) (*Request, error) {
	if t.responseListener == nil {
		err := errors.New("request tracker's response listener not active")
		log.WithFields(log.Fields{
			"error": err,
		}).Error(err)
		return nil, err
	}

	unixReq := req

	if req.ResponseHook.Scheme != "unix" {
		addr := t.responseListener.Addr().String()
		responseHook, err := url.ParseRequestURI(fmt.Sprintf("unix://%s", addr))
		if err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"address": addr,
			}).Error("unable to parse tracker response address")
			return nil, err
		}

		unixReq = &Request{
			ID:           req.ID,
			Task:         req.Task,
			ResponseHook: responseHook,
			Args:         req.Args,
			// Success and ErrorHandler are unnecessary here and intentionally
			// omitted.
		}

		t.TrackRequest(req)
	}

	return unixReq, nil
}
