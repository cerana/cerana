package simple

import (
	"sync"

	"github.com/mistifyio/acomm"
)

// MultiRequest provides a way to manage multiple parallel requests
type MultiRequest struct {
	idsToNames map[string]string
	respWG     sync.WaitGroup
	responses  chan *acomm.Response
	tracker    *acomm.Tracker
}

// NewMultiRequest creates and initializes a new MultiRequest.
func NewMultiRequest(tracker *acomm.Tracker) *MultiRequest {
	return &MultiRequest{
		idsToNames: make(map[string]string),
		responses:  make(chan *acomm.Response, 100),
		tracker:    tracker,
	}
}

// AddRequest adds a request to the MultiRequest. Sending the request is still
// the responsibility of the caller.
func (m *MultiRequest) AddRequest(name string, req *acomm.Request) error {
	m.idsToNames[req.ID] = name

	req.SuccessHandler = m.responseHandler
	req.ErrorHandler = m.responseHandler

	m.respWG.Add(1)
	return m.tracker.TrackRequest(req)
}

// RemoveRequest removes a request from the MultiRequest. Useful if the send fails.
func (m *MultiRequest) RemoveRequest(req *acomm.Request) {
	if m.tracker.RemoveRequest(req) {
		m.respWG.Done()
	}
}

// responseHandler alerts the MultiRequest when a response has been received and
// captures the response.
func (m *MultiRequest) responseHandler(req *acomm.Request, resp *acomm.Response) {
	defer m.respWG.Done()

	m.responses <- resp
}

// Results returns responses for all of the requests, keyed on the request name
// (as opposed to request id). Blocks until all requests are accounted for.
func (m *MultiRequest) Responses() map[string]*acomm.Response {
	results := make(map[string]*acomm.Response)

	m.respWG.Wait()

	close(m.responses)
	for resp := range m.responses {
		name := m.idsToNames[resp.ID]
		results[name] = resp
	}

	return results
}
