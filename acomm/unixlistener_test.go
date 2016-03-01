package acomm_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"

	"github.com/mistifyio/mistify/acomm"
	"github.com/stretchr/testify/suite"
)

type UnixListenerTestSuite struct {
	suite.Suite
	Listener *acomm.UnixListener
	Socket   string
}

func (s *UnixListenerTestSuite) SetupTest() {
	f, err := ioutil.TempFile("", "acommTest-")
	s.Require().NoError(err, "failed to create temp socket")
	s.Require().NoError(f.Close(), "failed to close temp socket file")
	s.Require().NoError(os.Remove(f.Name()), "failed to remove temp socket file")

	s.Socket = fmt.Sprintf("%s.sock", f.Name())
	s.Listener = acomm.NewUnixListener(s.Socket, 0)
}

func (s *UnixListenerTestSuite) TearDownTest() {
	s.Listener.Stop(0)
	sConn := s.Listener.NextConn()
	s.Nil(sConn)
}

func TestUnixListenerTestSuite(t *testing.T) {
	suite.Run(t, new(UnixListenerTestSuite))
}

func (s *UnixListenerTestSuite) TestNewUnixListener() {
	s.NotNil(acomm.NewUnixListener("foobar", 0), "should have returned a new UnixListener")
	_ = s.Listener.Start()
}

func (s *UnixListenerTestSuite) TestAddr() {
	s.Equal(s.Socket, s.Listener.Addr(), "should be the same addr")
	_ = s.Listener.Start()
}

func (s *UnixListenerTestSuite) TestURL() {
	s.Equal(fmt.Sprintf("unix://%s", s.Socket), s.Listener.URL().String(), "should be the same URL")
	_ = s.Listener.Start()
}

func (s *UnixListenerTestSuite) TestStart() {
	s.NoError(s.Listener.Start(), "should start successfully")
	s.Error(s.Listener.Start(), "should error calling start again")

	bad := acomm.NewUnixListener(s.Listener.Addr(), 0)
	s.Error(bad.Start(), "should not be able to start on a used socket")
}

func (s *UnixListenerTestSuite) TestNextAndDoneConn() {
	if !s.NoError(s.Listener.Start(), "should start successfully") {
		return
	}

	addr, _ := net.ResolveUnixAddr("unix", s.Listener.Addr())
	conn, err := net.DialUnix("unix", nil, addr)
	if !s.NoError(err, "failed to dial listener") {
		return
	}
	_, _ = conn.Write([]byte("foobar"))
	_ = conn.Close()

	lConn := s.Listener.NextConn()
	if !s.NotNil(lConn, "connection should not be nil") {
		return
	}

	s.Listener.DoneConn(lConn)
}

func (s *UnixListenerTestSuite) TestSendAndUnmarshalConnData() {
	if !s.NoError(s.Listener.Start(), "should start successfully") {
		return
	}

	in := map[string]string{
		"foo": "bar",
	}

	addr, _ := net.ResolveUnixAddr("unix", s.Listener.Addr())
	conn, err := net.DialUnix("unix", nil, addr)
	if !s.NoError(err, "failed to dial listener") {
		return
	}
	_ = acomm.SendConnData(conn, in)
	_ = conn.Close()

	lConn := s.Listener.NextConn()
	if !s.NotNil(lConn, "connection should not be nil") {
		return
	}

	out := map[string]string{}
	s.NoError(acomm.UnmarshalConnData(lConn, &out), "should succeed unmarshalling")

	s.Equal(in, out, "should have unmarshaled the correct data")

	s.Listener.DoneConn(lConn)
}
