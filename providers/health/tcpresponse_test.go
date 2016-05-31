package health_test

import (
	"fmt"
	"io"
	"net"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/logrusx"
	healthp "github.com/cerana/cerana/providers/health"
)

func (s *health) TestTCPResponse() {
	listener, err := tcpEchoer()
	s.Require().NoError(err)
	defer logrusx.LogReturnedErr(listener.Close, nil, "failed to close test listener")
	addr := listener.Addr().String()

	tests := []struct {
		address     string
		body        string
		regexp      string
		expectedErr string
	}{
		{"", "foo", "foo", "missing arg: address"},
		{addr, "foo", "", "missing arg: regexp"},
		{addr, "foo", "asdf", "response did not match"},
		{addr, "foo", "[", "error parsing regexp: missing closing ]: `[`"},
		{addr, "foo", "foo", ""},
		{addr, "foobarbaz", "bar", ""},
		{addr, "foobarbaz", "^bar$", "response did not match"},
	}

	for _, test := range tests {
		desc := fmt.Sprintf("%+v", test)
		args := &healthp.TCPResponseArgs{
			Address: test.address,
			Regexp:  test.regexp,
		}
		if test.body != "" {
			args.Body = []byte(test.body)
		}
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task:         "health-tcp-response",
			ResponseHook: s.responseHook,
			Args:         args,
		})
		s.Require().NoError(err, desc)

		resp, stream, err := s.health.TCPResponse(req)
		s.Nil(resp, desc)
		s.Nil(stream, desc)
		if test.expectedErr == "" {
			s.Nil(err, desc)
		} else {
			s.EqualError(err, test.expectedErr, desc)
		}
	}
}

func tcpEchoer() (net.Listener, error) {
	l, err := net.Listen("tcp", ":0")

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				break
			}

			go func(c net.Conn) {
				_ = io.Copy(c, c)
				_ = c.Close()
			}(conn)
		}
	}()

	return l, err
}
