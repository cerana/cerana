package acomm_test

import (
	"errors"
	"fmt"
	"net/url"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
)

type RequestTestSuite struct {
	suite.Suite
}

func (s *RequestTestSuite) SetupSuite() {
	log.SetLevel(log.FatalLevel)
}

func TestRequestTestSuite(t *testing.T) {
	suite.Run(t, new(RequestTestSuite))
}

func (s *RequestTestSuite) TestNewRequest() {
	tests := []struct {
		description string
		task        string
	}{
		{"without task", ""},
		{"with task", "foobar"},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		req := acomm.NewRequest(test.task)
		if !s.NotNil(req, msg("should have returned a request")) {
			continue
		}

		s.NotEmpty(req.ID, msg("should have set an ID"))
		s.Equal(test.task, req.Task, msg("should have set the task"))
	}
}

func (s *RequestTestSuite) TestSetResponseHook() {
	tests := []struct {
		description  string
		responseHook string
		expectedErr  bool
	}{
		{"empty", "", true},
		{"invalid", "asdf", true},
		{"unix", "unix://asdf", false},
		{"http", "http://asdf", false},
		{"https", "https://asdf", false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		req := &acomm.Request{}
		err := req.SetResponseHook(test.responseHook)
		if test.expectedErr {
			s.Error(err, msg("should have errored"))
			s.Nil(req.ResponseHook, msg("should not have set response hook"))
		} else {
			s.NoError(err, msg("should not have errored"))
			s.NotNil(req.ResponseHook, msg("should have set response hook"))
			s.Equal(test.responseHook, req.ResponseHook.String(), msg("should be equivalent response hooks"))
		}
	}
}

func (s *RequestTestSuite) TestSetStreamURL() {
	tests := []struct {
		description string
		streamURL   string
		expectedErr bool
	}{
		{"empty", "", true},
		{"invalid", "asdf", true},
		{"unix", "unix://asdf", false},
		{"http", "http://asdf", false},
		{"https", "https://asdf", false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		req := &acomm.Request{}
		err := req.SetStreamURL(test.streamURL)
		if test.expectedErr {
			s.Error(err, msg("should have errored"))
			s.Nil(req.StreamURL, msg("should not have set stream url"))
		} else {
			s.NoError(err, msg("should not have errored"))
			s.NotNil(req.StreamURL, msg("should have set stream url"))
			s.Equal(test.streamURL, req.StreamURL.String(), msg("should be equivalent stream urls"))
		}
	}
}

func (s *RequestTestSuite) TestSetTaskURL() {
	tests := []struct {
		description string
		taskURL     string
		expectedErr bool
	}{
		{"empty", "", true},
		{"invalid", "asdf", true},
		{"unix", "unix://asdf", false},
		{"http", "http://asdf", false},
		{"https", "https://asdf", false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		req := &acomm.Request{}
		err := req.SetTaskURL(test.taskURL)
		if test.expectedErr {
			s.Error(err, msg("should have errored"))
			s.Nil(req.TaskURL, msg("should not have set task url"))
		} else {
			s.NoError(err, msg("should not have errored"))
			s.NotNil(req.TaskURL, msg("should have set task url"))
			s.Equal(test.taskURL, req.TaskURL.String(), msg("should be equivalent task urls"))
		}
	}
}

func (s *RequestTestSuite) TestSetArgs() {
	tests := []struct {
		description string
		args        interface{}
	}{
		{"nil", nil},
		{"map", map[string]string{"foo": "bar"}},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		req := &acomm.Request{}
		s.NoError(req.SetArgs(test.args), msg("should not error"))
		s.NotNil(req.Args, msg("should have set args"))
	}
}

func (s *RequestTestSuite) TestHandleResponse() {
	sh, eh, handled := generateHandlers()
	respErr := errors.New("foobar")

	tests := []struct {
		description string
		sh          acomm.ResponseHandler
		eh          acomm.ResponseHandler
		respResult  interface{}
		respErr     error
	}{
		{"sh handler, err resp", sh, nil, nil, respErr},
		{"sh handler, success resp", sh, nil, struct{}{}, nil},
		{"eh handler, err resp", nil, eh, nil, respErr},
		{"eh handler, success resp", nil, eh, struct{}{}, nil},
		{"both handlers, err resp", sh, eh, nil, respErr},
		{"both handlers, success resp", sh, eh, struct{}{}, nil},
	}

	for _, test := range tests {
		handled["success"] = 0
		handled["error"] = 0
		msg := testMsgFunc(test.description)
		req := acomm.NewRequest("foobar")
		_ = req.SetArgs(struct{}{})
		req.SuccessHandler = test.sh
		req.ErrorHandler = test.eh

		resp, err := acomm.NewResponse(req, test.respResult, nil, test.respErr)
		if !s.NoError(err, msg("should not fail to build resp")) {
			continue
		}

		req.HandleResponse(resp)
		if test.respErr != nil {
			s.Equal(0, handled["success"], msg("should not have called success handler"))
			if test.eh != nil {
				s.Equal(1, handled["error"], msg("should have called error handler"))
			} else {
				s.Equal(0, handled["error"], msg("should not have called error handler"))
			}
		} else {
			if test.sh != nil {
				s.Equal(1, handled["success"], msg("should have called success handler"))
			} else {
				s.Equal(0, handled["success"], msg("should not have called success handler"))
			}
			s.Equal(0, handled["error"], msg("should not have called error handler"))
		}
	}
}

func (s *RequestTestSuite) TestValidate() {
	rh, _ := url.ParseRequestURI("unix:///tmp")

	tests := []struct {
		description  string
		id           string
		task         string
		responseHook *url.URL
		expectedErr  bool
	}{
		{"missing ID", "", "foo", rh, true},
		{"missing Task", uuid.New(), "", rh, true},
		{"missing ResponseHook", uuid.New(), "foo", nil, false},
		{"valid", uuid.New(), "foo", rh, false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		req := &acomm.Request{
			ID:           test.id,
			Task:         test.task,
			ResponseHook: test.responseHook,
		}
		if test.expectedErr {
			s.Error(req.Validate(), msg("should not be valid"))
		} else {
			s.NoError(req.Validate(), msg("should be valid"))
		}
	}
}

func generateHandlers() (acomm.ResponseHandler, acomm.ResponseHandler, map[string]int) {
	handled := make(map[string]int)
	sh := func(req *acomm.Request, resp *acomm.Response) {
		handled["success"]++
	}
	eh := func(req *acomm.Request, resp *acomm.Response) {
		handled["error"]++
	}
	return sh, eh, handled
}

func testMsgFunc(prefix string) func(...interface{}) string {
	return func(val ...interface{}) string {
		if len(val) == 0 {
			return prefix
		}
		msgPrefix := prefix + " : "
		if len(val) == 1 {
			return msgPrefix + val[0].(string)
		} else {
			return msgPrefix + fmt.Sprintf(val[0].(string), val[1:]...)
		}
	}
}
