package coordinator_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/coordinator"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
)

func TestServer(t *testing.T) {
	suite.Run(t, new(ServerSuite))
}

type ServerSuite struct {
	suite.Suite
	config     *coordinator.Config
	configData *coordinator.ConfigData
	server     *coordinator.Server
}

type params struct {
	ID string
}

func (s *ServerSuite) SetupSuite() {
	log.SetLevel(log.FatalLevel)

	socketDir, err := ioutil.TempDir("", "coordinatorTest-")
	s.Require().NoError(err, "failed to create socket dir")

	s.configData = &coordinator.ConfigData{
		SocketDir:      socketDir,
		ServiceName:    uuid.New(),
		ExternalPort:   45678,
		RequestTimeout: 5,
		LogLevel:       "fatal",
	}

	s.config, _, _, _, err = newConfig(true, false, s.configData)
	s.Require().NoError(err, "failed to create config")
	s.Require().NoError(s.config.LoadConfig(), "failed to load config")
}

func (s *ServerSuite) SetupTest() {
	var err error
	s.server, err = coordinator.NewServer(s.config)
	s.Require().NoError(err, "failed to create server")
	s.Require().NotNil(s.server, "failed to create server")
}

func (s *ServerSuite) TearDownSuite() {
	_ = os.RemoveAll(s.configData.SocketDir)
}

func (s *ServerSuite) TestNewServer() {
	configInvalid := coordinator.NewConfig(nil, nil)

	server, err := coordinator.NewServer(configInvalid)
	s.Nil(server, "should not create server with invalid config")
	s.Error(err, "should error with invalid config")
}

func (s *ServerSuite) TestReqRespHandle() {
	// Start
	if !s.NoError(s.server.Start(), "failed to start server") {
		return
	}
	time.Sleep(time.Second)
	// Stop
	defer s.server.Stop()

	// Set up handlers
	result := make(chan *params, 10)

	// Task handler
	taskName := "foobar"
	taskListener := s.createTaskListener(taskName, result)
	if taskListener == nil {
		return
	}
	defer taskListener.Stop(0)

	// Response handlers
	responseServer, responseListener := s.createResponseHandlers(result)
	if responseServer != nil {
		defer responseServer.Close()
	}
	if responseListener != nil {
		defer responseListener.Stop(0)
	}
	if responseServer == nil || responseListener == nil {
		return
	}

	// Coordinator URLs
	internalURL, _ := url.ParseRequestURI("unix://" + filepath.Join(
		s.config.SocketDir(),
		"coordinator",
		s.config.ServiceName()+".sock"),
	)
	externalURL, _ := url.ParseRequestURI(fmt.Sprintf(
		"http://localhost:%v",
		s.configData.ExternalPort),
	)

	// Test cases
	tests := []struct {
		description  string
		taskName     string
		internal     bool
		params       *params
		expectFailed bool
	}{
		{"valid http", taskName, false, &params{uuid.New()}, false},
		{"valid unix", taskName, true, &params{uuid.New()}, false},
		{"bad task http", "asdf", false, &params{uuid.New()}, true},
		{"bad task unix", "asdf", true, &params{uuid.New()}, true},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		hookURL := responseServer.URL
		coordinatorURL := externalURL
		if test.internal {
			hookURL = responseListener.URL().String()
			coordinatorURL = internalURL
		}

		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task:               test.taskName,
			ResponseHookString: hookURL,
			Args:               test.params,
		})
		s.Require().NoError(err, msg("should have created req"))

		if err := acomm.Send(coordinatorURL, req); err != nil {
			result <- nil
		}

		respData := <-result
		if test.expectFailed {
			s.Nil(respData, msg("should have failed"))
		} else {
			s.Equal(test.params, respData, msg("should have gotten the correct response data"))
		}

		drainChan(result)
	}
}

func (s *ServerSuite) TestStopOnSignal() {
	selfProcess, err := os.FindProcess(os.Getpid())
	if !s.NoError(err, "couldn't find this process") {
		return
	}

	if !s.NoError(s.server.Start(), "failed to start server") {
		return
	}

	stopSignal := syscall.SIGUSR1

	done := make(chan struct{})
	go func() {
		s.server.StopOnSignal(stopSignal)
		close(done)
	}()

	time.Sleep(time.Second)
	_ = selfProcess.Signal(stopSignal)

	<-done
}

func (s *ServerSuite) createTaskListener(taskName string, result chan *params) *acomm.UnixListener {
	taskListener := acomm.NewUnixListener(filepath.Join(s.configData.SocketDir, taskName, "test.sock"), 0)
	if !s.NoError(taskListener.Start(), "failed to start task listener") {
		return nil
	}

	go func() {
		for {
			conn := taskListener.NextConn()
			if conn == nil {
				break
			}
			defer taskListener.DoneConn(conn)
			req := &acomm.Request{}
			if err := acomm.UnmarshalConnData(conn, req); err != nil {
				result <- nil
				continue
			}

			params := &params{}
			_ = req.UnmarshalArgs(params)

			// Respond to the initial request
			resp, _ := acomm.NewResponse(req, nil, nil, nil)
			if err := acomm.SendConnData(conn, resp); err != nil {
				result <- nil
				continue
			}

			// Response to hook
			resp, _ = acomm.NewResponse(req, req.Args, nil, nil)
			if err := req.Respond(resp); err != nil {
				result <- nil
				continue
			}
		}
	}()

	time.Sleep(time.Second)
	return taskListener
}

func (s *ServerSuite) createResponseHandlers(result chan *params) (*httptest.Server, *acomm.UnixListener) {
	// HTTP response
	responseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := &acomm.Response{}
		decoder := json.NewDecoder(r.Body)
		_ = decoder.Decode(resp)

		p := &params{}
		_ = resp.UnmarshalResult(p)
		result <- p
	}))

	// Unix response
	responseListener := acomm.NewUnixListener(filepath.Join(s.configData.SocketDir, "testResponse.sock"), 0)
	if !s.NoError(responseListener.Start(), "failed to start task listener") {
		return responseServer, nil
	}

	go func() {
		conn := responseListener.NextConn()
		if conn == nil {
			return
		}
		defer responseListener.DoneConn(conn)

		resp := &acomm.Response{}
		_ = acomm.UnmarshalConnData(conn, resp)
		p := &params{}
		_ = resp.UnmarshalResult(p)
		result <- p
	}()

	return responseServer, responseListener
}

// drainChan is a helper function to drain a channel, such as between test cases
func drainChan(ch chan *params) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
