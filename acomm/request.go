package acomm

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/cerana/cerana/pkg/errors"
	"github.com/pborman/uuid"
)

// Request is a request data structure for asynchronous requests. The ID is
// used to identify the request throught its life cycle. The ResponseHook is a
// URL where response data should be sent. SuccessHandler and ErrorHandler will
// be called appropriately to handle a response.
type Request struct {
	ID             string           `json:"id"`
	Task           string           `json:"task"`
	TaskURL        *url.URL         `json:"taskURL"`
	ResponseHook   *url.URL         `json:"responseHook"`
	StreamURL      *url.URL         `json:"streamURL"`
	Args           *json.RawMessage `json:"args"`
	SuccessHandler ResponseHandler  `json:"-"`
	ErrorHandler   ResponseHandler  `json:"-"`
	timeout        *time.Timer
	proxied        bool
}

// RequestOptions are properties and options used to create a new Request
// object. There are options to either directly specify a URL or provide a
// string that will be parsed.
type RequestOptions struct {
	Task               string
	TaskURL            *url.URL
	TaskURLString      string
	ResponseHook       *url.URL
	ResponseHookString string
	StreamURL          *url.URL
	StreamURLString    string
	Args               interface{}
	SuccessHandler     ResponseHandler `json:"-"`
	ErrorHandler       ResponseHandler `json:"-"`
}

// ResponseHandler is a function to run when a request receives a response.
type ResponseHandler func(*Request, *Response)

// NewRequest creates a new Request instance.
func NewRequest(opts RequestOptions) (*Request, error) {
	req := &Request{
		ID:             uuid.New(),
		Task:           opts.Task,
		SuccessHandler: opts.SuccessHandler,
		ErrorHandler:   opts.ErrorHandler,
	}

	if err := req.SetArgs(opts.Args); err != nil {
		return nil, err
	}

	if opts.TaskURL != nil {
		req.TaskURL = opts.TaskURL
	} else if opts.TaskURLString != "" {
		if err := req.SetTaskURL(opts.TaskURLString); err != nil {
			return nil, err
		}
	}

	if opts.ResponseHook != nil {
		req.ResponseHook = opts.ResponseHook
	} else if opts.ResponseHookString != "" {
		if err := req.SetResponseHook(opts.ResponseHookString); err != nil {
			return nil, err
		}
	}

	if opts.StreamURL != nil {
		req.StreamURL = opts.StreamURL
	} else if opts.StreamURLString != "" {
		if err := req.SetStreamURL(opts.StreamURLString); err != nil {
			return nil, err
		}
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	return req, nil
}

// SetResponseHook is a convenience method to set the ResponseHook from a
// string url.
func (req *Request) SetResponseHook(urlString string) error {
	responseHook, err := url.ParseRequestURI(urlString)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"requestID": req.ID, "responseHook": urlString})
	}

	req.ResponseHook = responseHook
	return nil
}

// SetStreamURL is a convenience method to set the StreamURL from a string url.
func (req *Request) SetStreamURL(urlString string) error {
	streamURL, err := url.ParseRequestURI(urlString)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"requestID": req.ID, "streamURL": urlString})
	}

	req.StreamURL = streamURL
	return nil
}

// SetTaskURL is a convenience method to set the TaskURL from a string url.
func (req *Request) SetTaskURL(urlString string) error {
	taskURL, err := url.ParseRequestURI(urlString)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"requestID": req.ID, "taskURL": urlString})
	}

	req.TaskURL = taskURL
	return nil
}

// SetArgs sets the Args.
func (req *Request) SetArgs(args interface{}) error {
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"requestID": req.ID, "args": args})
	}
	req.Args = (*json.RawMessage)(&argsJSON)
	return nil
}

// UnmarshalArgs unmarshals the request args into the destination object.
func (req *Request) UnmarshalArgs(dest interface{}) error {
	return errors.Wrapv(unmarshalFromRaw(req.Args, dest), map[string]interface{}{"requestID": req.ID})
}

func unmarshalFromRaw(src *json.RawMessage, dest interface{}) error {
	if src == nil {
		return nil
	}

	if err := json.Unmarshal(*src, dest); err != nil {
		return errors.Wrapv(err, map[string]interface{}{"json": string(*src)})
	}
	return nil
}

// Validate validates the reqeust
func (req *Request) Validate() error {
	if req.ID == "" {
		return errors.Newv("missing id", map[string]interface{}{"request": req})
	}
	if req.Task == "" {
		return errors.Newv("missing task", map[string]interface{}{"requestID": req.ID, "request": req})
	}

	return nil
}

// Respond sends a Response to the ResponseHook if present.
func (req *Request) Respond(resp *Response) error {
	if req.ResponseHook == nil {
		return nil
	}
	return errors.Wrapv(Send(req.ResponseHook, resp), map[string]interface{}{"requestID": req.ID})
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
