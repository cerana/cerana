package logrusx_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/stretchr/testify/suite"
)

type FormatterTestSuite struct {
	suite.Suite
	Log    *logrus.Logger
	Buffer *bytes.Buffer
}

func (s *FormatterTestSuite) SetupTest() {
	s.Buffer = new(bytes.Buffer)

	// Setup a new Logger instance
	s.Log = logrus.New()
	s.Log.Out = s.Buffer
	s.Log.Formatter = &logrusx.JSONFormatter{}
}

func TestFormatterTestSuite(t *testing.T) {
	suite.Run(t, new(FormatterTestSuite))
}

func (s *FormatterTestSuite) TestJSONFormatterFormat() {
	testErr := errors.New("test error message")
	entry := &logrus.Entry{
		Logger: s.Log,
		Data: logrus.Fields{
			"error": testErr,
		},
		Time:    time.Now(),
		Level:   s.Log.Level,
		Message: "test error message",
	}

	mf := &logrusx.JSONFormatter{}
	jsonBytes, err := mf.Format(entry)
	s.NoError(err)
	var loggedEntry map[string]interface{}
	s.NoError(json.Unmarshal(jsonBytes, &loggedEntry))

	errMap := loggedEntry["error"].(map[string]interface{})
	s.Equal(errMap["Message"], testErr.Error())
}

func (s *FormatterTestSuite) TestJSONFormatterWithLogrus() {
	fieldNames := []string{"error", "asdf"}
	testErr := errors.New("test error message")

	for _, fieldName := range fieldNames {
		s.Log.WithField(fieldName, errors.New("test error message")).Info("test info message")
		var entry map[string]interface{}
		s.NoError(json.Unmarshal(s.Buffer.Bytes(), &entry))
		s.Buffer.Truncate(0)

		errMap := entry[fieldName].(map[string]interface{})
		s.Equal(errMap["Message"], testErr.Error())
	}
}
