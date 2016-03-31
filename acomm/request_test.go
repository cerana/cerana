package acomm_test

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
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
	task := "foobar"
	args := map[string]string{
		"foo": "bar",
	}

	sh, eh, _ := generateHandlers()

	tests := []struct {
		description  string
		task         string
		responseHook string
		streamURL    string
		args         interface{}
		sh           acomm.ResponseHandler
		eh           acomm.ResponseHandler
		expectedErr  bool
	}{
		{"missing response hook", task, "", "", args, sh, eh, true},
		{"invalid response hook", task, "asdf", "", args, sh, eh, true},
		{"missing args", task, "unix://asdf", "", nil, sh, eh, false},
		{"unix hook and args", task, "unix://asdf", "", args, sh, eh, false},
		{"http hook and args", task, "http://asdf", "", args, sh, eh, false},
		{"https hook and args", task, "https://asdf", "", args, sh, eh, false},
		{"unix stream", task, "unix://asdf", "unix://asdf", args, sh, eh, false},
		{"http stream", task, "http://asdf", "http://asdf", args, sh, eh, false},
		{"https stream", task, "https://asdf", "https://asdf", args, sh, eh, false},
		{"unix hook, args, no handlers", task, "unix://asdf", "", args, nil, nil, false},
		{"unix hook, args, sh handler", task, "unix://asdf", "", args, sh, nil, false},
		{"unix hook, args, eh handler", task, "unix://asdf", "", args, nil, eh, false},
		{"missing task ", "", "unix://asdf", "", args, sh, eh, true},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		req, err := acomm.NewRequest(test.task, test.responseHook, test.streamURL, test.args, test.sh, test.eh)
		if test.expectedErr {
			s.Error(err, msg("should have failed"))
			s.Nil(req, msg("should not have returned a request"))
		} else {
			if !s.NoError(err, msg("should have succeeded")) {
				s.T().Log(msg(err.Error()))
				continue
			}
			if !s.NotNil(req, msg("should have returned a request")) {
				continue
			}

			s.NotEmpty(req.ID, msg("should have set an ID"))
			s.Equal(test.task, req.Task, msg("should have set the task"))
			s.Equal(test.responseHook, req.ResponseHook.String(), msg("should have set the response hook"))
			if test.streamURL != "" {
				s.Equal(test.streamURL, req.StreamURL.String(), msg("should have set the stream url"))
			}
			var args map[string]string
			s.NoError(req.UnmarshalArgs(&args))
			if test.args == nil {
				s.Nil(args, msg("should have nil arguments"))
			} else {
				s.Equal(test.args, args, msg("should have set the arguments"))
			}
			s.Equal(reflect.ValueOf(test.sh).Pointer(), reflect.ValueOf(req.SuccessHandler).Pointer(), msg("should have set success handler"))
			s.Equal(reflect.ValueOf(test.eh).Pointer(), reflect.ValueOf(req.ErrorHandler).Pointer(), msg("should have set error handler"))
		}
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
		req, err := acomm.NewRequest("foobar", "unix://foo", "", struct{}{}, test.sh, test.eh)
		if !s.NoError(err, msg("should not fail to build req")) {
			continue
		}
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
		{"missing ResponseHook", uuid.New(), "foo", nil, true},
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
