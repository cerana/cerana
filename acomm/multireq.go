package acomm

import (
	"sync"
	"time"
)

// MultiRequest provides a way to manage multiple parallel requests
type MultiRequest struct {
	idsToNames map[string]string
	respWG     sync.WaitGroup
	responses  chan *Response
	tracker    *Tracker
	timeout    time.Duration
}

// NewMultiRequest creates and initializes a new MultiRequest.
func NewMultiRequest(tracker *Tracker, timeout time.Duration) *MultiRequest {
	return &MultiRequest{
		idsToNames: make(map[string]string),
		responses:  make(chan *Response, 100),
		tracker:    tracker,
		timeout:    timeout,
	}
}

// AddRequest adds a request to the MultiRequest. Sending the request is still
// the responsibility of the caller.
func (m *MultiRequest) AddRequest(name string, req *Request) error {
	m.idsToNames[req.ID] = name
	req.ResponseHook = m.tracker.URL()
	req.SuccessHandler = m.responseHandler
	req.ErrorHandler = m.responseHandler

	m.respWG.Add(1)
	return m.tracker.TrackRequest(req, m.timeout)
}

// RemoveRequest removes a request from the MultiRequest. Useful if the send fails.
func (m *MultiRequest) RemoveRequest(req *Request) {
	if m.tracker.RemoveRequest(req) {
		m.respWG.Done()
	}
}

// responseHandler alerts the MultiRequest when a response has been received and
// captures the response.
func (m *MultiRequest) responseHandler(req *Request, resp *Response) {
	defer m.respWG.Done()

	m.responses <- resp
}

// Responses returns responses for all of the requests, keyed on the request name
// (as opposed to request id). Blocks until all requests are accounted for.
func (m *MultiRequest) Responses() map[string]*Response {
	results := make(map[string]*Response)

	m.respWG.Wait()

	close(m.responses)
	for resp := range m.responses {
		name := m.idsToNames[resp.ID]
		results[name] = resp
	}

	return results
}
