package coordinator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/tylerb/graceful"
)

// Server is the coordinator server. It handles accepting internal and external
// requests and proxying them to appropriate providers.
type Server struct {
	config   *Config
	proxy    *acomm.Tracker
	internal *acomm.UnixListener
	external *graceful.Server
}

// NewServer creates and initializes a new instance of Server.
func NewServer(config *Config) (*Server, error) {
	http.DefaultTransport.(*http.Transport).DisableKeepAlives = true
	if err := config.Validate(); err != nil {
		return nil, err
	}

	var err error
	s := &Server{
		config: config,
	}

	// Internal socket for requests from providers
	internalSocket := filepath.Join(
		config.SocketDir(),
		"coordinator",
		config.ServiceName()+".sock")
	s.internal = acomm.NewUnixListener(internalSocket, 0)

	// Response socket for proxied requests
	responseSocket := filepath.Join(
		config.SocketDir(),
		"response",
		config.ServiceName()+".sock")

	streamURLS := fmt.Sprintf("http://localhost:%d/stream", config.ExternalPort())
	streamURL, err := url.ParseRequestURI(streamURLS)
	if err != nil {
		return nil, errors.Wrapv(err, map[string]interface{}{
			"externalPort": config.ExternalPort(),
			"streamURL":    streamURLS,
		}, "failed to generate valid streamURL")
	}
	proxyURLS := fmt.Sprintf("http://localhost:%d/proxy", config.ExternalPort())
	proxyURL, err := url.ParseRequestURI(proxyURLS)
	if err != nil {
		return nil, errors.Wrapv(err, map[string]interface{}{
			"externalPort": config.ExternalPort(),
			"proxyURL":     proxyURLS,
		}, "failed to generate valid proxyURL")
	}
	s.proxy, err = acomm.NewTracker(responseSocket, streamURL, proxyURL, config.RequestTimeout())
	if err != nil {
		return nil, err
	}

	// External server for requests to and from outside
	mux := http.NewServeMux()
	mux.HandleFunc("/stream", acomm.ProxyStreamHandler)
	mux.HandleFunc("/proxy", s.proxy.ProxyExternalHandler)
	mux.HandleFunc("/", s.externalHandler)
	s.external = &graceful.Server{
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", config.ExternalPort()),
			Handler: mux,
		},
		NoSignalHandling: true,
	}

	logrus.WithFields(logrus.Fields{
		"response": responseSocket,
		"stream":   streamURL.String(),
		"internal": internalSocket,
		"external": fmt.Sprintf(":%d", config.ExternalPort()),
	}).Info("server addresses")

	return s, nil
}

// externalHandler is the http handler for external requests.
func (s *Server) externalHandler(w http.ResponseWriter, r *http.Request) {
	var respErr error
	req := &acomm.Request{}

	// Send the immediate response
	defer func() {
		resp, err := acomm.NewResponse(req, nil, nil, respErr)
		errData := map[string]interface{}{
			"request":  req,
			"response": resp,
		}
		respJSON, err := json.Marshal(resp)
		if err != nil {
			err = errors.Wrapv(err, errData)
			logrus.WithField("error", err).Error("failed to marshal initial response")
		}

		if _, err := w.Write(respJSON); err != nil {
			err = errors.Wrapv(err, errData)
			logrus.WithField("error", err).Error("failed to send initial response")
		}
	}()

	// Parse the request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respErr = errors.Wrap(err, "failed to read request body")
		return
	}

	if err := json.Unmarshal(body, req); err != nil {
		respErr = errors.Wrapv(err, map[string]interface{}{"json": string(body)}, "failed to unmarshal request")
		return
	}

	if err := acomm.ReplaceLocalhost(req.ResponseHook, r.RemoteAddr); err != nil {
		respErr = errors.Wrapv(err, map[string]interface{}{"request": req}, "responseHook")
		return
	}
	if err := acomm.ReplaceLocalhost(req.StreamURL, r.RemoteAddr); err != nil {
		respErr = errors.Wrapv(err, map[string]interface{}{"request": req}, "streamURL")
		return
	}

	respErr = s.handleRequest(req)
}

func (s *Server) internalHandler() {
	for {
		conn := s.internal.NextConn()
		if conn == nil {
			return
		}
		go s.acceptInternalRequest(conn)
	}
}

func (s *Server) acceptInternalRequest(conn net.Conn) {
	defer s.internal.DoneConn(conn)
	var respErr error
	req := &acomm.Request{}
	defer func() {
		// Respond to the initial request
		resp, err := acomm.NewResponse(req, nil, nil, respErr)
		errData := map[string]interface{}{
			"request":  req,
			"response": resp,
		}
		if err != nil {
			err = errors.Wrapv(err, errData)
			logrus.WithField("error", err).Error("failed to marshal initial response")
			return
		}

		if err := acomm.SendConnData(conn, resp); err != nil {
			err = errors.Wrapv(err, errData)
			logrus.WithField("error", err).Error("failed to send initial response")
			return
		}
	}()

	if err := acomm.UnmarshalConnData(conn, req); err != nil {
		respErr = errors.Wrap(err, "failed to unmarshal request")
		return
	}

	if err := req.Validate(); err != nil {
		respErr = errors.Wrapv(err, map[string]interface{}{"request": req})
		return
	}

	respErr = s.handleRequest(req)
}

func (s *Server) handleRequest(req *acomm.Request) error {
	var err error
	if req.TaskURL == nil {
		err = s.localTask(req)
	} else {
		err = s.externalTask(req)
	}
	if err != nil {
		_ = s.proxy.RemoveRequest(req)
	}
	return errors.Wrapv(err, map[string]interface{}{"request": req})
}

// localTask handles proxying and forwarding a request to a provider for
// the specified task.
func (s *Server) localTask(req *acomm.Request) error {
	providerSockets, err := s.getProviders(req.Task)
	if err != nil {
		return err
	}

	if len(providerSockets) == 0 {
		return errors.Newv("no providers available for task", map[string]interface{}{"task": req.Task})
	}

	proxyReq, err := s.proxy.ProxyUnix(req, 0)
	if err != nil {
		return err
	}

	// Cycle through available providers until one accepts the request
	for _, providerSocket := range providerSockets {
		addr, _ := url.ParseRequestURI(fmt.Sprintf("unix://%s", providerSocket))
		err = acomm.Send(addr, proxyReq)
		if err == nil {
			// Successfully sent
			break
		}
	}

	return err
}

// externalTask handles proxying and forwarding a request to an external
// service (e.g. another coordinator)
func (s *Server) externalTask(req *acomm.Request) error {
	taskURL := req.TaskURL
	proxyReq := req
	if taskURL.Scheme != "unix" {
		var err error
		proxyReq, err = s.proxy.ProxyExternal(req, 0)
		if err != nil {
			return err
		}
	} else {
		// Don't proxy local requests
		proxyReq.TaskURL = nil
	}
	return acomm.Send(taskURL, proxyReq)
}

// getProviders returns a list of providers registered for a given task.
func (s *Server) getProviders(task string) ([]string, error) {
	// Find Task Providers
	if task == "" {
		return nil, errors.New("request missing task")
	}

	taskSocketDir := filepath.Join(s.config.SocketDir(), task)
	files, err := ioutil.ReadDir(taskSocketDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrapv(err, map[string]interface{}{"task": task, "taskSocketDir": taskSocketDir})
	}

	// Filter out any non-socket files
	providerSockets := make([]string, 0, len(files))
	for _, fi := range files {
		if fi.Mode()&os.ModeSocket == os.ModeSocket {
			providerSockets = append(providerSockets, filepath.Join(taskSocketDir, fi.Name()))
		}
	}

	return providerSockets, nil
}

// externalListenAndServe runs and blocks on the external http server.
func (s *Server) externalListenAndServe() {
	if err := s.external.ListenAndServe(); err != nil {
		// Ignore the error from closing the listener, which is involved in the
		// graceful shutdown
		if !strings.Contains(err.Error(), "use of closed network connection") {
			logrus.WithField("error", errors.Wrap(err)).Error("server error")

			// Stop the coordinator if this was unexpected
			s.Stop()
		}
	}
}

// Start starts the server, running all of the listeners and proxy tracker.
func (s *Server) Start() error {
	// Start up the proxy tracker
	if err := s.proxy.Start(); err != nil {
		return err
	}

	// Start up the internal request handler
	if err := s.internal.Start(); err != nil {
		return err
	}
	go s.internalHandler()

	// Start up the external request handler
	go s.externalListenAndServe()
	return nil
}

// Stop stops the server, gracefully stopping all of the listeners and proxy
// tracker.
func (s *Server) Stop() {
	// Stop accepting new external requests
	stopChan := s.external.StopChan()
	s.external.Stop(0)
	<-stopChan

	// Stop accepting new internal requests
	s.internal.Stop(0)

	// Stop the proxy tracker
	s.proxy.Stop()
}

// StopOnSignal will wait until one of the specified signals is received and
// then stop the server. If no signals are specified, it will use a default
// set.
func (s *Server) StopOnSignal(signals ...os.Signal) {
	if len(signals) == 0 {
		signals = []os.Signal{os.Interrupt, os.Kill, syscall.SIGTERM}
	}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, signals...)
	sig := <-sigChan
	logrus.WithFields(logrus.Fields{
		"signal": sig,
	}).Info("signal received, stopping")

	s.Stop()
}
