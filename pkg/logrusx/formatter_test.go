package logrusx_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	logx "github.com/mistifyio/mistify-logrus-ext"
	"github.com/stretchr/testify/suite"
)

type FormatterTestSuite struct {
	suite.Suite
	Log    *log.Logger
	Buffer *bytes.Buffer
}

func (s *FormatterTestSuite) SetupTest() {
	s.Buffer = new(bytes.Buffer)

	// Setup a new Logger instance
	s.Log = log.New()
	s.Log.Out = s.Buffer
	s.Log.Formatter = &logx.MistifyFormatter{}
}

func TestFormatterTestSuite(t *testing.T) {
	suite.Run(t, new(FormatterTestSuite))
}

func (s *FormatterTestSuite) TestMistifyFormatterFormat() {
	testErr := errors.New("test error message")
	entry := &log.Entry{
		Logger: s.Log,
		Data: log.Fields{
			"error": testErr,
		},
		Time:    time.Now(),
		Level:   s.Log.Level,
		Message: "test error message",
	}

	mf := &logx.MistifyFormatter{}
	jsonBytes, err := mf.Format(entry)
	s.NoError(err)
	var loggedEntry map[string]interface{}
	s.NoError(json.Unmarshal(jsonBytes, &loggedEntry))

	errMap := loggedEntry["error"].(map[string]interface{})
	s.Equal(errMap["Message"], testErr.Error())
}

func (s *FormatterTestSuite) TestMistifyFormatterWithLogrus() {
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
