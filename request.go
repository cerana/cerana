package acomm

import (
	"errors"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

// Request is a request data structure for asynchronous requests. The ID is
// used to identify the request throught its life cycle. The ResponseHook is a
// URL where response data should be sent. SuccessHandler and ErrorHandler will
// be called appropriately to handle a response.
type Request struct {
	ID             string          `json:"id"`
	Task           string          `json:"task"`
	ResponseHook   *url.URL        `json:"responsehook"`
	Args           interface{}     `json:"args"`
	SuccessHandler ResponseHandler `json:"-"`
	ErrorHandler   ResponseHandler `json:"-"`
	proxied        bool
}

// ResponseHandler is a function to run when a request receives a response.
type ResponseHandler func(*Request, *Response)

// NewRequest creates a new Request instance.
func NewRequest(task, responseHook string, args interface{}, sh ResponseHandler, eh ResponseHandler) (*Request, error) {
	hook, err := url.ParseRequestURI(responseHook)
	if err != nil {
		log.WithFields(log.Fields{
			"error":        err,
			"responseHook": responseHook,
		}).Error("invalid response hook url")
		return nil, err
	}

	if task == "" {
		return nil, errors.New("missing task")
	}

	return &Request{
		ID:             uuid.New(),
		Task:           task,
		ResponseHook:   hook,
		Args:           args,
		SuccessHandler: sh,
		ErrorHandler:   eh,
	}, nil
}

// Validate validates the reqeust
func (req *Request) Validate() error {
	if req.ID == "" {
		err := errors.New("missing id")
		log.WithFields(log.Fields{
			"req":   req,
			"error": err,
		}).Error("invalid req")
		return err
	}
	if req.Task == "" {
		err := errors.New("missing task")
		log.WithFields(log.Fields{
			"req":   req,
			"error": err,
		}).Error("invalid req")
		return err
	}
	if req.ResponseHook == nil {
		err := errors.New("missing response hook")
		log.WithFields(log.Fields{
			"req":   req,
			"error": err,
		}).Error("invalid req")
		return err
	}
	return nil
}

// Respond sends a Response to a Request's ResponseHook.
func (req *Request) Respond(resp *Response) error {
	return Send(req.ResponseHook, resp)
}

// HandleResponse determines whether a response indicates success or error and
// runs the appropriate handler. If the appropriate handler is not defined, it
// is assumed no handling is necessary and silently finishes.
func (req *Request) HandleResponse(resp *Response) {
	if resp.Error != nil {
		if req.ErrorHandler != nil {
			req.ErrorHandler(req, resp)
		}
		return
	}

	if req.SuccessHandler != nil {
		req.SuccessHandler(req, resp)
		return
	}
}
