package acomm_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/async-comm"
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
	args := map[string]string{
		"foo": "bar",
	}

	sh, eh, _ := generateHandlers()

	tests := []struct {
		description  string
		responseHook string
		args         interface{}
		sh           acomm.ResponseHandler
		eh           acomm.ResponseHandler
		expectedErr  bool
	}{
		{"missing response hook", "", args, sh, eh, true},
		{"invalid response hook", "asdf", args, sh, eh, true},
		{"missing args", "unix://asdf", nil, sh, eh, false},
		{"unix hook and args", "unix://asdf", args, sh, eh, false},
		{"http hook and args", "http://asdf", args, sh, eh, false},
		{"https hook and args", "https://asdf", args, sh, eh, false},
		{"unix hook, args, no handlers", "unix://asdf", args, nil, nil, false},
		{"unix hook, args, sh handler", "unix://asdf", args, sh, nil, false},
		{"unix hook, args, eh handler", "unix://asdf", args, nil, eh, false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		req, err := acomm.NewRequest(test.responseHook, test.args, test.sh, test.eh)
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
			s.Equal(test.responseHook, req.ResponseHook.String(), msg("should have set the response hook"))
			s.Equal(test.args, req.Args, msg("should have set the arguments"))
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
		req, err := acomm.NewRequest("unix://foo", struct{}{}, test.sh, test.eh)
		if !s.NoError(err, msg("should not fail to build req")) {
			continue
		}
		resp, err := acomm.NewResponse(req, test.respResult, test.respErr)
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
