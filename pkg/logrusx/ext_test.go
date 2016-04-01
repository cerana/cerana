package logrusx_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	log "github.com/Sirupsen/logrus"
	logx "github.com/mistifyio/mistify-logrus-ext"
	"github.com/stretchr/testify/suite"
)

type ExtTestSuite struct {
	suite.Suite
	DefaultFormatter log.Formatter
	DefaultLevel     log.Level
	DefaultOut       io.Writer
}

func (s *ExtTestSuite) restoreDefaults() {
	// Restore logrus defaults
	log.SetFormatter(s.DefaultFormatter)
	log.SetLevel(s.DefaultLevel)
	log.SetOutput(s.DefaultOut)
}

func (s *ExtTestSuite) SetupSuite() {
	// Save logrus defaults
	std := log.StandardLogger()
	s.DefaultFormatter = std.Formatter
	s.DefaultLevel = std.Level
	s.DefaultOut = std.Out
}

func (s *ExtTestSuite) SetupTest() {
	// Make sure each test starts with defaults
	s.restoreDefaults()
}

func (s *ExtTestSuite) TearDownSuite() {
	// Make sure defaults are restored for other test suites
	s.restoreDefaults()
}

func TestExtTestSuite(t *testing.T) {
	suite.Run(t, new(ExtTestSuite))
}

func (s *ExtTestSuite) TestSetLevel() {
	// Bad value
	s.Error(logx.SetLevel("foobar"))
	s.Equal(s.DefaultLevel, log.GetLevel())

	// Good value
	s.NoError(logx.SetLevel("debug"))
	s.Equal(log.DebugLevel, log.GetLevel())
}

func (s *ExtTestSuite) TestDefaultSetup() {
	// Bad Value
	s.Error(logx.DefaultSetup("foobar"))
	s.Equal(s.DefaultLevel, log.GetLevel())

	// Good Value
	std := log.StandardLogger()
	s.NoError(logx.DefaultSetup("debug"))
	s.Equal(log.DebugLevel, log.GetLevel())
	s.IsType(&logx.MistifyFormatter{}, std.Formatter)
}

func (s *ExtTestSuite) TestLogReturnedErr() {
	var buffer bytes.Buffer
	var out string
	log.SetOutput(&buffer)

	logMsg := "qwerty"
	errMsg := "foobar"

	// No error logged
	returnsNil := func() error {
		return nil
	}

	logx.LogReturnedErr(returnsNil, nil, logMsg)
	s.Empty(buffer.String())

	// Error logged
	returnsError := func() error {
		return errors.New(errMsg)
	}

	logx.LogReturnedErr(returnsError, nil, logMsg)
	out = buffer.String()
	s.Contains(out, logMsg)
	s.Contains(out, errMsg)
}
