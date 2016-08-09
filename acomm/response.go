package acomm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/logrusx"
)

// Response is a response data structure for asynchronous requests. The ID
// should be the same as the Request it corresponds to. Result should be nil if
// Error is present and vice versa.
type Response struct {
	ID        string           `json:"id"`
	Result    *json.RawMessage `json:"result"`
	StreamURL *url.URL         `json:"streamURL"`
	Error     error            `json:"error"`
}

// MarshalJSON marshals a Response into JSON.
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

// UnmarshalJSON unmarshals JSON data into a Response.
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
func NewResponse(req *Request, result interface{}, streamURL *url.URL, respErr error) (*Response, error) {
	if req == nil {
		err := errors.New("cannot create response without request")
		logrus.WithFields(logrus.Fields{
			"errors": err,
		}).Error(err)
		return nil, err
	}

	if result != nil && respErr != nil {
		err := errors.New("cannot set both result and err")
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error(err)
		return nil, err
	}

	var resultRaw *json.RawMessage
	if result != nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":  err,
				"result": result,
			}).Error("failed to marshal response result")
		}
		resultRaw = (*json.RawMessage)(&resultJSON)
	}

	return &Response{
		ID:        req.ID,
		Result:    resultRaw,
		Error:     respErr,
		StreamURL: streamURL,
	}, nil
}

// UnmarshalResult unmarshals the response result into the destination object.
func (r *Response) UnmarshalResult(dest interface{}) error {
	return unmarshalFromRaw(r.Result, dest)
}

// Send attempts send the payload to the specified URL.
func Send(addr *url.URL, payload interface{}) error {
	if addr == nil {
		err := errors.New("missing addr")
		logrus.WithFields(logrus.Fields{
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
		logrus.WithFields(logrus.Fields{
			"error": err,
			"type":  addr.Scheme,
			"addr":  addr,
		}).Error("cannot not send to url")
		return err
	}
}

// sendUnix sends a request or response via a Unix socket.
func sendUnix(addr *url.URL, payload interface{}) error {
	conn, err := net.Dial("unix", addr.RequestURI())
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"addr":    addr,
			"payload": payload,
		}).Error("failed to connect to unix socket")
		return err
	}
	defer logrusx.LogReturnedErr(conn.Close,
		logrus.Fields{"addr": addr},
		"failed to close unix connection",
	)

	if err := SendConnData(conn, payload); err != nil {
		return err
	}

	resp := &Response{}
	if err := UnmarshalConnData(conn, resp); err != nil {
		return err
	}

	return resp.Error
}

// sendHTTP sends the Response via HTTP/HTTPS
func sendHTTP(addr *url.URL, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"payload": payload,
		}).Error("failed to marshal payload json")
		return err
	}

	httpResp, err := http.Post(addr.String(), "application/json", bytes.NewReader(payloadJSON))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"addr":    addr,
			"payload": payload,
		}).Error("failed to send payload")
		return err
	}
	defer logrusx.LogReturnedErr(httpResp.Body.Close, nil, "failed to close http body")

	body, _ := ioutil.ReadAll(httpResp.Body)
	resp := &Response{}
	if err := json.Unmarshal(body, resp); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"body":  string(body),
		}).Error(err)
		return err
	}

	return resp.Error
}
