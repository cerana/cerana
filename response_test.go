package acomm_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/async-comm"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
)

type ResponseTestSuite struct {
	suite.Suite
}

func (s *ResponseTestSuite) SetupSuite() {
	log.SetLevel(log.FatalLevel)
}

func TestResponseTestSuite(t *testing.T) {
	suite.Run(t, new(ResponseTestSuite))
}

func (s *ResponseTestSuite) TestNewResponse() {
	result := map[string]string{
		"foo": "bar",
	}

	request, _ := acomm.NewRequest("unix://foo", nil)
	respErr := errors.New("foobar")

	tests := []struct {
		description string
		request     *acomm.Request
		result      interface{}
		err         error
		expectedErr bool
	}{
		{"missing request", nil, result, nil, true},
		{"missing result and error", request, nil, nil, true},
		{"result and error", request, result, respErr, true},
		{"result only", request, result, nil, false},
		{"error only", request, nil, respErr, false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		resp, err := acomm.NewResponse(test.request, test.result, test.err)
		if test.expectedErr {
			s.Error(err, msg("should have failed"))
			s.Nil(resp, msg("should not have returned a response"))
		} else {
			if !s.NoError(err, msg("should have succeeded")) {
				s.T().Log(msg(err.Error()))
				continue
			}
			if !s.NotNil(resp, msg("should have returned a response")) {
				continue
			}
			s.Equal(test.request.ID, resp.ID, msg("should have set an ID"))
			s.Equal(test.result, resp.Result, msg("should have set the result"))
			s.Equal(test.err, resp.Error, msg("should have set the error"))
		}
	}
}

func (s *ResponseTestSuite) TestSend() {
	sentResp := &acomm.Response{}

	// Mock HTTP response server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		s.NoError(err, "should not fail reading body")
		s.NoError(json.Unmarshal(body, sentResp), "should not fail unmarshalling response")
	}))
	defer ts.Close()

	// Mock Unix response listener
	f, err := ioutil.TempFile("", "acommTest-")
	s.Require().NoError(err, "failed to create test unix socket")
	_ = f.Close()
	_ = os.Remove(f.Name())
	socketPath := fmt.Sprintf("%s.sock", f.Name())
	listener, err := net.Listen("unix", socketPath)
	s.Require().NoError(err, "failed to listen on unix socket")
	defer func() { _ = listener.Close() }()
	go func() {
		for {
			conn, err := listener.Accept()
			s.Require().NoError(err, "listener accept error")
			body, err := ioutil.ReadAll(conn)
			s.NoError(err, "should not fail reading body")
			s.NoError(json.Unmarshal(body, sentResp), "should not fail unmarshalling response")
			conn.Close()
		}
	}()

	response := &acomm.Response{
		ID:     uuid.New(),
		Result: map[string]string{"foo": "bar"},
	}

	tests := []struct {
		responseHook string
		expectedErr  bool
	}{
		{ts.URL, false},
		{"http://badpath", true},
		{fmt.Sprintf("unix://%s", socketPath), false},
		{fmt.Sprintf("unix://%s", "badpath"), true},
		{"foobar://", true},
	}

	for _, test := range tests {
		sentResp = &acomm.Response{}
		u, _ := url.ParseRequestURI(test.responseHook)
		err := response.Send(u)
		time.Sleep(10 * time.Millisecond)
		if test.expectedErr {
			s.Error(err, test.responseHook)
			s.Empty(sentResp.ID, test.responseHook)
		} else {
			if !s.NoError(err, test.responseHook) {
				continue
			}
			s.Equal(response.ID, sentResp.ID, test.responseHook)
		}
	}

}
