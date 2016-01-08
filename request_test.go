package acomm_test

import (
	"fmt"
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

	tests := []struct {
		description  string
		responseHook string
		args         interface{}
		expectedErr  bool
	}{
		{"missing response hook", "", args, true},
		{"invalid response hook", "asdf", args, true},
		{"missing args", "unix://asdf", nil, false},
		{"unix hook and args", "unix://asdf", args, false},
		{"http hook and args", "http://asdf", args, false},
		{"https hook and args", "https://asdf", args, false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		req, err := acomm.NewRequest(test.responseHook, test.args)
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
		}
	}
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
