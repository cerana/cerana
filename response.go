package acomm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
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

func (r *Response) MarshalJSON() ([]byte, error) {
	type Alias Response
	respErr := r.Error
	if respErr == nil {
		respErr = errors.New("")
	}
	return json.Marshal(&struct {
		Error string `json:"error"`
		*Alias
	}{
		Error: respErr.Error(),
		Alias: (*Alias)(r),
	})
}

func (r *Response) UnmarshalJSON(data []byte) error {
	type Alias Response
	aux := &struct {
		Error string `json:"error"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Error != "" {
		r.Error = errors.New(aux.Error)
	}
	return nil
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

	return &Response{
		ID:     req.ID,
		Result: result,
		Error:  err,
	}, nil
}

// Send attempts send the payload to the specified URL.
func Send(addr *url.URL, payload interface{}) error {
	if addr == nil {
		err := errors.New("missing addr")
		log.WithFields(log.Fields{
			"error": err,
		}).Error(err)
		return err
	}
	switch addr.Scheme {
	case "unix":
		return sendUnix(addr, payload)
	case "http", "https":
		return sendHTTP(addr, payload)
	default:
		err := errors.New("unknown url type")
		log.WithFields(log.Fields{
			"error": err,
			"type":  addr.Scheme,
			"addr":  addr,
		}).Error(err)
		return err
	}
}

// sendUnix sends a request or response via a Unix socket.
func sendUnix(addr *url.URL, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"payload": payload,
		}).Error("failed to marshal payload json")
		return err
	}

	conn, err := net.Dial("unix", addr.RequestURI())
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"addr":    addr,
			"payload": payload,
		}).Error("failed to connect to unix socket")
		return err
	}
	defer logx.LogReturnedErr(conn.Close,
		log.Fields{"addr": addr},
		"failed to close unix connection",
	)

	if _, err := conn.Write(payloadJSON); err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"addr":    addr,
			"payload": payload,
		}).Error("failed to write payload to unix socket")
		return err
	}
	return nil
}

// sendHTTP sends the Response via HTTP/HTTPS
func sendHTTP(addr *url.URL, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"payload": payload,
		}).Error("failed to marshal payload json")
		return err
	}

	httpResp, err := http.Post(addr.String(), "application/json", bytes.NewReader(payloadJSON))
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"addr":    addr,
			"payload": payload,
		}).Error("failed to send payload")
		return err
	}
	defer httpResp.Body.Close()

	body, _ := ioutil.ReadAll(httpResp.Body)
	resp := &Response{}
	if err := json.Unmarshal(body, resp); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"body":  string(body),
		}).Error(err)
		return err
	}

	return resp.Error
}
