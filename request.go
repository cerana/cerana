package acomm

import (
	"errors"
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
	hook, err := url.Parse(responseHook)
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

// Respond sends a Response to a Request's ResponseHook. Supports UNIX sockets
// and http/https URLs.
func (req *Request) Respond(resp *Response) error {
	switch req.ResponseHook.Scheme {
	case "unix":
		return resp.SendUnix(req.ResponseHook.String())
	case "http", "https":
		return resp.SendHTTP(req.ResponseHook.String())
	default:
		err := errors.New("unknown response hook type")
		log.WithFields(log.Fields{
			"error":        err,
			"type":         req.ResponseHook.Scheme,
			"responseHook": req.ResponseHook,
		}).Error(err)
		return err
	}
}
