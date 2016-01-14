package acomm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
)

const (
	statusStarted = iota
	statusStopping
	statusStopped
)

// Tracker keeps track of requests waiting on a response.
type Tracker struct {
	status           int
	responseListener *UnixListener
	requestsLock     sync.Mutex // Protects requests
	requests         map[string]*Request
	waitgroup        sync.WaitGroup
}

// NewTracker creates and initializes a new Tracker. If a socketDir is not
// provided, the response socket will be created in a temporary directory.
func NewTracker(socketPath string) (*Tracker, error) {
	if socketPath == "" {
		var err error
		socketPath, err = generateTempSocketPath()
		if err != nil {
			return nil, err
		}
	}

	return &Tracker{
		status:           statusStopped,
		responseListener: NewUnixListener(socketPath),
	}, nil
}

func generateTempSocketPath() (string, error) {
	// Use TempFile to allocate a uniquely named file in either the specified
	// dir or the default temp dir. It is then removed so that the unix socket
	// can be created with that name.
	f, err := ioutil.TempFile("", "acommTrackerResponses-")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("failed to create temp file for response socket")
		return "", err
	}
	_ = f.Close()
	_ = os.Remove(f.Name())

	return fmt.Sprintf("%s.sock", f.Name()), nil
}

// NumRequests returns the number of tracked requests
func (t *Tracker) NumRequests() int {
	t.requestsLock.Lock()
	defer t.requestsLock.Unlock()

	return len(t.requests)
}

// Addr returns the string representation of the Tracker's response listener socket.
func (t *Tracker) Addr() string {
	return t.responseListener.Addr()
}

// URL returns the URL of the Tracker's response listener socket.
func (t *Tracker) URL() *url.URL {
	return t.responseListener.URL()
}

// Start activates the tracker. This allows tracking of requests as well as
// listening for and handling responses.
func (t *Tracker) Start() error {
	if t.status == statusStarted {
		return nil
	}
	if t.status == statusStopping {
		return errors.New("can't start tracker while stopping")
	}

	t.requests = make(map[string]*Request)

	// start the proxy response listener
	if err := t.responseListener.Start(); err != nil {
		return err
	}

	go t.listenForResponses()

	t.status = statusStarted

	return nil
}

// listenForResponse continually accepts new responses on the listener.
func (t *Tracker) listenForResponses() {
	for {
		conn := t.responseListener.NextConn()
		if conn == nil {
			return
		}

		go t.handleConn(conn)
	}
}

// handleConn handles the response connection and parses the data.
func (t *Tracker) handleConn(conn net.Conn) {
	defer t.responseListener.DoneConn(conn)

	resp := &Response{}
	if err := UnmarshalConnData(conn, resp); err != nil {
		return
	}

	go t.HandleResponse(resp)
}

// HandleResponse associates a response with a request and either forwards the
// response or calls the request's handler.
func (t *Tracker) HandleResponse(resp *Response) {
	req := t.retrieveRequest(resp.ID)
	if req == nil {
		err := errors.New("response does not have tracked request")
		log.WithFields(log.Fields{
			"error":    err,
			"response": resp,
		}).Error(err)
		return
	}
	defer t.waitgroup.Done()

	// If there are handlers, this is the final destination, so handle the
	// response. Otherwise, forward the response along.
	// Known issue: If this is the final destination and there are
	// no handlers, there will be an extra redirects back here. Since the
	// request has already been removed from the tracker, it will only happen
	// once.
	if !req.proxied {
		req.HandleResponse(resp)
		return
	}

	// Forward the response along
	_ = req.Respond(resp)
	return
}

// Stop deactivates the tracker. It blocks until all active connections or tracked requests to finish.
func (t *Tracker) Stop() {
	// Nothing to do if it's not listening.
	if t.responseListener == nil {
		return
	}

	// Prevent new requests from being tracked
	t.status = statusStopping

	// Handle any requests that are expected
	t.waitgroup.Wait()
	// Stop listening for new requests
	t.responseListener.Stop()
	t.status = statusStopped
	return
}

// TrackRequest tracks a request. This does not need to be called after using
// ProxyUnix.
func (t *Tracker) TrackRequest(req *Request) error {
	t.requestsLock.Lock()
	defer t.requestsLock.Unlock()

	if t.status == statusStarted {
		if _, ok := t.requests[req.ID]; ok {
			err := errors.New("request id already traacked")
			log.WithFields(log.Fields{
				"request": req,
				"error":   err,
			}).Error(err)
			return err
		}
		t.waitgroup.Add(1)
		t.requests[req.ID] = req
		return nil
	}

	err := errors.New("failed to track request in unstarted tracker")
	log.WithFields(log.Fields{
		"request":       req,
		"trackerStatus": t.status,
		"error":         err,
	}).Error(err)
	return err
}

// RemoveRequest should be used to remove a tracked request. Use in cases such
// as sending failures, where there is no hope of a response being received.
func (t *Tracker) RemoveRequest(req *Request) bool {
	if r := t.retrieveRequest(req.ID); r != nil {
		t.waitgroup.Done()
		return true
	}
	return false
}

// retrieveRequest returns a tracked Request based on ID and stops tracking it.
func (t *Tracker) retrieveRequest(id string) *Request {
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
		unixReq = &Request{
			ID:           req.ID,
			Task:         req.Task,
			ResponseHook: t.responseListener.URL(),
			Args:         req.Args,
			// Success and ErrorHandler are unnecessary here and intentionally
			// omitted.
		}
		if err := t.TrackRequest(req); err != nil {
			return nil, err
		}
		req.proxied = true
	}

	return unixReq, nil
}
