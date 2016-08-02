package logrusx_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/stretchr/testify/suite"
)

type ExtTestSuite struct {
	suite.Suite
	DefaultFormatter logrus.Formatter
	DefaultLevel     logrus.Level
	DefaultOut       io.Writer
}

func (s *ExtTestSuite) restoreDefaults() {
	// Restore logrus defaults
	logrus.SetFormatter(s.DefaultFormatter)
	logrus.SetLevel(s.DefaultLevel)
	logrus.SetOutput(s.DefaultOut)
}

func (s *ExtTestSuite) SetupSuite() {
	// Save logrus defaults
	std := logrus.StandardLogger()
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
	s.Error(logrusx.SetLevel("foobar"))
	s.Equal(s.DefaultLevel, logrus.GetLevel())

	// Good value
	s.NoError(logrusx.SetLevel("debug"))
	s.Equal(logrus.DebugLevel, logrus.GetLevel())
}

func (s *ExtTestSuite) TestDefaultSetup() {
	// Bad Value
	s.Error(logrusx.DefaultSetup("foobar"))
	s.Equal(s.DefaultLevel, logrus.GetLevel())

	// Good Value
	std := logrus.StandardLogger()
	s.NoError(logrusx.DefaultSetup("debug"))
	s.Equal(logrus.DebugLevel, logrus.GetLevel())
	s.IsType(&logrusx.JSONFormatter{}, std.Formatter)
}

func (s *ExtTestSuite) TestLogReturnedErr() {
	var buffer bytes.Buffer
	var out string
	logrus.SetOutput(&buffer)

	logMsg := "qwerty"
	errMsg := "foobar"

	// No error logged
	returnsNil := func() error {
		return nil
	}

	logrusx.LogReturnedErr(returnsNil, nil, logMsg)
	s.Empty(buffer.String())

	// Error logged
	returnsError := func() error {
		return errors.New(errMsg)
	}

	logrusx.LogReturnedErr(returnsError, nil, logMsg)
	out = buffer.String()
	s.Contains(out, logMsg)
	s.Contains(out, errMsg)
}
