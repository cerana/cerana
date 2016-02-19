package provider_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ServerSuite struct {
	suite.Suite
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerSuite))
}

func (s *ServerSuite) TestNewServer() {
}

func (s *ServerSuite) TestTracker() {
}

func (s *ServerSuite) TestRegisterTask() {
}

func (s *ServerSuite) TestRegisteredTasks() {
}

func (s *ServerSuite) TestStart() {
}

func (s *ServerSuite) TestStop() {
}

func (s *ServerSuite) TestStopOnSignal() {
}
