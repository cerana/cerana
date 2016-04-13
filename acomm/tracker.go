package acomm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/logrusx"
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
	httpStreamURL    *url.URL
	externalProxyURL *url.URL
	defaultTimeout   time.Duration
	requestsLock     sync.Mutex // Protects requests
	requests         map[string]*Request
	dsLock           sync.Mutex // Protects dataStreams
	dataStreams      map[string]*UnixListener
	waitgroup        sync.WaitGroup
}

// NewTracker creates and initializes a new Tracker. If a socketPath is not
// provided, the response socket will be created in a temporary directory.
func NewTracker(socketPath string, httpStreamURL, externalProxyURL *url.URL, defaultTimeout time.Duration) (*Tracker, error) {
	if socketPath == "" {
		var err error
		socketPath, err = generateTempSocketPath("", "acommTrackerResponses-")
		if err != nil {
			return nil, err
		}
	}

	if defaultTimeout <= 0 {
		defaultTimeout = time.Minute
	}

	return &Tracker{
		status:           statusStopped,
		responseListener: NewUnixListener(socketPath, 0),
		httpStreamURL:    httpStreamURL,
		externalProxyURL: externalProxyURL,
		dataStreams:      make(map[string]*UnixListener),
		defaultTimeout:   defaultTimeout,
	}, nil
}

func generateTempSocketPath(dir, prefix string) (string, error) {
	// Use TempFile to allocate a uniquely named file in either the specified
	// dir or the default temp dir. It is then removed so that the unix socket
	// can be created with that name.
	// TODO: Decide on permissions
	if dir != "" {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			log.WithFields(log.Fields{
				"directory": dir,
				"perm":      os.ModePerm,
				"error":     err,
			}).Error("failed to create directory for socket")
			return "", err
		}
	}

	f, err := ioutil.TempFile(dir, prefix)
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

	_ = SendConnData(conn, &Response{})

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

	// Stop the request timeout. The result doesn't matter.
	if req.timeout != nil {
		_ = req.timeout.Stop()
	}

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

	if resp.StreamURL != nil {
		streamURL, err := t.ProxyStreamHTTPURL(resp.StreamURL) // Replace the StreamURL with a proxy stream url
		if err != nil {
			streamURL = nil
		}
		resp.StreamURL = streamURL
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

	// Stop listening for responses
	t.responseListener.Stop(0)

	// Stop any data streamers
	var dsWG sync.WaitGroup
	t.dsLock.Lock()
	for _, ds := range t.dataStreams {
		dsWG.Add(1)
		go func(ds *UnixListener) {
			defer dsWG.Done()
			ds.Stop(0)
		}(ds)
	}
	t.dsLock.Unlock()
	dsWG.Wait()

	t.status = statusStopped
	return
}

// TrackRequest tracks a request. This does not need to be called after using
// ProxyUnix.
func (t *Tracker) TrackRequest(req *Request, timeout time.Duration) error {
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

		t.setRequestTimeout(req, timeout)
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
		if r.timeout != nil {
			_ = r.timeout.Stop()
		}
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

func (t *Tracker) setRequestTimeout(req *Request, timeout time.Duration) {
	// Fallback to default timeout
	if timeout <= 0 {
		timeout = t.defaultTimeout
	}

	resp, err := NewResponse(req, nil, nil, errors.New("response timeout"))
	if err != nil {
		return
	}

	req.timeout = time.AfterFunc(timeout, func() {
		t.HandleResponse(resp)
	})
	return
}

// ProxyUnix proxies requests that have response hooks and stream urls of
// non-unix sockets. If the response hook and stream url are already unix
// sockets, it returns the original request. If the response hook is not, it
// tracks the original request and returns a new request with a unix socket
// response hook. If the stream url is not, it pipes the original stream
// through a new unix socket and updates the stream url. The purpose of this is
// so that there can be a single entry and exit point for external
// communication, while local services can reply directly to each other.
func (t *Tracker) ProxyUnix(req *Request, timeout time.Duration) (*Request, error) {
	if t.responseListener == nil {
		err := errors.New("request tracker's response listener not active")
		log.WithFields(log.Fields{
			"error": err,
		}).Error(err)
		return nil, err
	}

	if req.StreamURL != nil && req.StreamURL.Scheme != "unix" {
		// proxy the stream
		r, w := io.Pipe()

		go func(src *url.URL) {
			defer logrusx.LogReturnedErr(w.Close, nil, "failed to close proxy stream writer")
			if err := Stream(w, src); err != nil {
				log.WithFields(log.Fields{
					"error":     err,
					"streamURL": src,
				}).Error("failed to stream")
			}
		}(req.StreamURL)

		addr, err := t.NewStreamUnix("", r)
		if err != nil {
			return nil, err
		}

		req.StreamURL = addr
	}

	unixReq := req
	if req.ResponseHook.Scheme != "unix" {
		// proxy the request
		unixReq = &Request{
			ID:           req.ID,
			Task:         req.Task,
			ResponseHook: t.responseListener.URL(),
			StreamURL:    req.StreamURL,
			Args:         req.Args,
			// Success and ErrorHandler are unnecessary here and intentionally
			// omitted.
		}
		if err := t.TrackRequest(req, timeout); err != nil {
			return nil, err
		}
		req.proxied = true
	}

	return unixReq, nil
}

// ProxyExternal proxies a request intended for an external destination
func (t *Tracker) ProxyExternal(req *Request, timeout time.Duration) (*Request, error) {
	if t.externalProxyURL == nil {
		err := errors.New("tracker missing external proxy url")
		log.WithFields(log.Fields{
			"error": err,
		}).Error(err)
		return nil, err
	}
	if t.responseListener == nil {
		err := errors.New("request tracker's response listener not active")
		log.WithFields(log.Fields{
			"error": err,
		}).Error(err)
		return nil, err
	}

	externalReq := &Request{
		ID:           req.ID,
		Task:         req.Task,
		ResponseHook: t.externalProxyURL,
		Args:         req.Args,
	}
	if req.StreamURL != nil {
		streamURL, err := t.ProxyStreamHTTPURL(req.StreamURL) // Replace the StreamURL with a proxy stream url
		if err != nil {
			return nil, err
		}
		externalReq.StreamURL = streamURL
	}

	if err := t.TrackRequest(req, timeout); err != nil {
		return nil, err
	}
	req.proxied = true

	return externalReq, nil
}

// ProxyExternalHandler is an HTTP HandlerFunc for proxying an external request.
func (t *Tracker) ProxyExternalHandler(w http.ResponseWriter, r *http.Request) {
	resp := &Response{}
	body, err := ioutil.ReadAll(r.Body)
	if err == nil {
		err = json.Unmarshal(body, resp)
	}

	ack := &Response{
		Error: err,
	}
	ackJSON, err := json.Marshal(ack)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"ack":   ack,
		}).Error("failed to marshal ack")
		return
	}
	if _, err := w.Write(ackJSON); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"ack":   ack,
		}).Error("failed to ack response")
		return
	}

	if ack.Error != nil {
		return
	}
	t.HandleResponse(resp)
}
