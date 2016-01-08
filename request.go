package acomm

import (
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

// Request is a request data structure for asynchronous requests. The ID is
// used to identify the request throught its life cycle. The ResponseHook is a
// URL where response data should be sent.
type Request struct {
	ID           string      `json:"id"`
	ResponseHook *url.URL    `json:"responsehook"`
	Args         interface{} `json:"args"`
}

// NewRequest creates a new Request instance.
func NewRequest(responseHook string, args interface{}) (*Request, error) {
	hook, err := url.ParseRequestURI(responseHook)
	if err != nil {
		log.WithFields(log.Fields{
			"error":        err,
			"responseHook": responseHook,
		}).Error("invalid response hook url")
		return nil, err
	}

	return &Request{
		ID:           uuid.New(),
		ResponseHook: hook,
		Args:         args,
	}, nil
}

// Respond sends a Response to a Request's ResponseHook.
func (req *Request) Respond(resp *Response) error {
	return resp.Send(req.ResponseHook)
}
