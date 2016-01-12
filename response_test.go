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
	"github.com/mistifyio/acomm"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
)

type ResponseTestSuite struct {
	suite.Suite
	Responses chan *acomm.Response
}

func (s *ResponseTestSuite) SetupSuite() {
	log.SetLevel(log.FatalLevel)
	s.Responses = make(chan *acomm.Response, 10)
}

func TestResponseTestSuite(t *testing.T) {
	suite.Run(t, new(ResponseTestSuite))
}

func (s *ResponseTestSuite) NextResp() *acomm.Response {
	timeout := make(chan struct{}, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- struct{}{}
	}()

	var resp *acomm.Response
	select {
	case resp = <-s.Responses:
	case <-timeout:
	}
	return resp
}

func (s *ResponseTestSuite) TestNewResponse() {
	result := map[string]string{
		"foo": "bar",
	}

	request, _ := acomm.NewRequest("unix://foo", nil, nil, nil)
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
	// Mock HTTP response server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := &acomm.Response{}
		body, err := ioutil.ReadAll(r.Body)
		s.NoError(err, "should not fail reading body")
		s.NoError(json.Unmarshal(body, resp), "should not fail unmarshalling response")
		s.Responses <- resp
	}))
	defer ts.Close()

	// Mock Unix response listener
	f, err := ioutil.TempFile("", "acommTest-")
	if !s.NoError(err, "failed to create test unix socket") {
		return
	}
	_ = f.Close()
	_ = os.Remove(f.Name())
	socketPath := fmt.Sprintf("%s.sock", f.Name())
	listener, err := net.Listen("unix", socketPath)
	if !s.NoError(err, "failed to listen on unix socket") {
		return
	}
	defer func() { _ = listener.Close() }()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			resp := &acomm.Response{}
			body, err := ioutil.ReadAll(conn)
			s.NoError(err, "should not fail reading body")
			s.NoError(json.Unmarshal(body, resp), "should not fail unmarshalling response")
			s.Responses <- resp
			_ = conn.Close()
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
		msg := testMsgFunc(test.responseHook)
		u, _ := url.ParseRequestURI(test.responseHook)
		err := response.Send(u)
		resp := s.NextResp()
		if test.expectedErr {
			s.Error(err, msg("send should fail"))
			s.Nil(resp, msg("response hook should not receive a response"))
		} else {
			if !s.NoError(err, msg("send should not fail")) {
				continue
			}
			s.Equal(response.ID, resp.ID, msg("response should be what was sent"))
		}
	}
}
