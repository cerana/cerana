package health

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/logrusx"
)

// HTTPStatusArgs are arguments for HTTPStatus health checks.
type HTTPStatusArgs struct {
	URL        string `json:"url"`
	Method     string `json:"method"`
	Body       []byte `json:"body"`
	StatusCode int    `json:"statusCode"`
}

// HTTPStatus makes an HTTP request to the specified URL and compares the
// response status code to an expected status code.
func (h *Health) HTTPStatus(req *acomm.Request) (interface{}, *url.URL, error) {
	var args HTTPStatusArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.URL == "" {
		return nil, nil, errors.New("missing arg: url")
	}

	if args.StatusCode == 0 {
		args.StatusCode = http.StatusOK
	}

	httpReq, err := http.NewRequest(args.Method, args.URL, bytes.NewReader(args.Body))
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, nil, err
	}
	defer logrusx.LogReturnedErr(httpResp.Body.Close, nil, "failed to close resp body")

	if httpResp.StatusCode != args.StatusCode {
		err = errors.New("unexpected response status code")
	}

	return nil, nil, err
}
