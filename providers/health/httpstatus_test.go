package health_test

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/cerana/cerana/acomm"
	healthp "github.com/cerana/cerana/providers/health"
)

func (s *health) TestHTTPStatus() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		if r.Method == "FAIL" {
			status = http.StatusBadRequest
		} else {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				status = http.StatusBadRequest
			} else if len(body) > 0 {
				status = int(binary.BigEndian.Uint64(body))
			}
		}
		w.WriteHeader(status)
	}))
	defer ts.Close()

	tests := []struct {
		url         string
		method      string
		statusCode  int
		body        uint64
		expectedErr string
	}{
		{"", "GET", http.StatusOK, 0, "missing arg: url"},
		{ts.URL, "", http.StatusOK, 0, ""},
		{ts.URL, "GET", 0, 0, ""},
		{ts.URL, "GET", http.StatusOK, 0, ""},
		{ts.URL, "GET", http.StatusAccepted, http.StatusAccepted, ""},
		{ts.URL, "FAIL", http.StatusOK, 0, "unexpected response status code"},
	}

	for _, test := range tests {
		desc := fmt.Sprintf("%+v", test)
		args := &healthp.HTTPStatusArgs{
			URL:        test.url,
			Method:     test.method,
			StatusCode: test.statusCode,
		}
		if test.body != 0 {
			b := make([]byte, 8)
			binary.BigEndian.PutUint64(b, test.body)
			args.Body = b
		}

		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task:         "health-http-status",
			ResponseHook: s.responseHook,
			Args:         args,
		})
		s.Require().NoError(err, desc)

		resp, stream, err := s.health.HTTPStatus(req)
		s.Nil(resp, desc)
		s.Nil(stream, desc)
		if test.expectedErr == "" {
			s.Nil(err, desc)
		} else {
			s.EqualError(err, test.expectedErr, desc)
		}
	}
}
