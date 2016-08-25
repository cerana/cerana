package acomm

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"github.com/cerana/cerana/pkg/errors"
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
		return errors.Wrapv(err, map[string]interface{}{"requestID": r.ID})
	}
	if aux.Error != "" {
		r.Error = errors.Newv(aux.Error, map[string]interface{}{"requestID": r.ID})
	}
	return nil
}

// NewResponse creates a new Response instance based on a Request.
func NewResponse(req *Request, result interface{}, streamURL *url.URL, respErr error) (*Response, error) {
	if req == nil {
		return nil, errors.New("cannot create response without request")
	}

	if result != nil && respErr != nil {
		return nil, errors.Newv("cannot set both result and err", map[string]interface{}{"requestID": req.ID})
	}

	var resultRaw *json.RawMessage
	if result != nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return nil, errors.Wrapv(err, map[string]interface{}{"requestID": req.ID, "result": result})
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
		return errors.New("missing addr")
	}
	switch addr.Scheme {
	case "unix":
		return sendUnix(addr, payload)
	case "http", "https":
		return sendHTTP(addr, payload)
	default:
		return errors.Newv("unknown url scheme", map[string]interface{}{"addr": addr})
	}
}

// sendUnix sends a request or response via a Unix socket.
func sendUnix(addr *url.URL, payload interface{}) error {
	conn, err := net.Dial("unix", addr.RequestURI())
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"addr": addr, "payload": payload})
	}
	defer logrusx.LogReturnedErr(conn.Close,
		map[string]interface{}{"addr": addr},
		"failed to close unix connection",
	)

	if err := SendConnData(conn, payload); err != nil {
		return err
	}

	resp := &Response{}
	if err := UnmarshalConnData(conn, resp); err != nil {
		return err
	}

	return errors.ResetStack(resp.Error)
}

// sendHTTP sends the Response via HTTP/HTTPS
func sendHTTP(addr *url.URL, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"payload": payload})
	}

	httpResp, err := http.Post(addr.String(), "application/json", bytes.NewReader(payloadJSON))
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"addr": addr, "payload": payload})
	}
	defer logrusx.LogReturnedErr(httpResp.Body.Close, nil, "failed to close http body")

	body, _ := ioutil.ReadAll(httpResp.Body)
	resp := &Response{}
	if err := json.Unmarshal(body, resp); err != nil {
		return errors.Wrapv(err, map[string]interface{}{"body": string(body)})
	}

	return errors.ResetStack(resp.Error)
}
