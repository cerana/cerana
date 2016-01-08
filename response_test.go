package acomm_test

import (
	"errors"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/async-comm"
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
