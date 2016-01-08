package acomm

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
	logx "github.com/mistifyio/mistify-logrus-ext"
)

// Response is a response data structure for asynchronous requests. The ID
// should be the same as the Request it corresponds to. Result should be nil if
// Error is present and vice versa.
type Response struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result"`
	Error  error       `json:"error"`
}

// NewResponse creates a new Response instance based on a Request.
func NewResponse(req *Request, result interface{}, err error) (*Response, error) {
	if req == nil {
		err := errors.New("cannot create response without request")
		log.WithFields(log.Fields{
			"errors": err,
		}).Error(err)
		return nil, err
	}

	if result != nil && err != nil {
		err := errors.New("cannot set both result and err")
		log.WithFields(log.Fields{
			"errors": err,
		}).Error(err)
		return nil, err
	}

	if result == nil && err == nil {
		err := errors.New("must set one of either result or err")
		log.WithFields(log.Fields{
			"errors": err,
		}).Error(err)
		return nil, err
	}

	return &Response{
		ID:     req.ID,
		Result: result,
		Error:  err,
	}, nil
}

func (resp *Response) Send(responseHook *url.URL) error {
	switch responseHook.Scheme {
	case "unix":
		return resp.sendUnix(responseHook)
	case "http", "https":
		return resp.sendHTTP(responseHook)
	default:
		err := errors.New("unknown response hook type")
		log.WithFields(log.Fields{
			"error":        err,
			"type":         responseHook.Scheme,
			"responseHook": responseHook,
		}).Error(err)
		return err
	}
}

// sendUnix sends the Response via a Unix socket.
func (resp *Response) sendUnix(responseHook *url.URL) error {
	respJSON, err := json.Marshal(resp)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"response": resp,
		}).Error("failed to marshal response json")
		return err
	}

	conn, err := net.Dial("unix", responseHook.RequestURI())
	if err != nil {
		log.WithFields(log.Fields{
			"error":        err,
			"responseHook": responseHook,
			"resp":         resp,
		}).Error("failed to connect to unix socket")
		return err
	}
	defer logx.LogReturnedErr(conn.Close,
		log.Fields{"responseHook": responseHook},
		"failed to close unix connection",
	)

	if _, err := conn.Write(respJSON); err != nil {
		log.WithFields(log.Fields{
			"error":        err,
			"responseHook": responseHook,
			"resp":         resp,
		}).Error("failed to connect to unix socket")
		return err
	}
	return nil
}

// sendHTTP sends the Response via HTTP/HTTPS
func (resp *Response) sendHTTP(responseHook *url.URL) error {
	respJSON, err := json.Marshal(resp)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"response": resp,
		}).Error("failed to marshal response json")
		return err
	}

	httpResp, err := http.Post(responseHook.String(), "application/json", bytes.NewReader(respJSON))
	if err != nil {
		log.WithFields(log.Fields{
			"error":        err,
			"responseHook": responseHook,
			"resp":         resp,
		}).Error("failed to respond to request")
		return err
	}

	if httpResp.StatusCode != http.StatusOK {
		err := errors.New("unexpected http code for request response")
		log.WithFields(log.Fields{
			"error":        err,
			"responseHook": responseHook,
			"resp":         resp,
			"code":         httpResp.StatusCode,
		}).Error(err)
		return err
	}
	return nil
}
