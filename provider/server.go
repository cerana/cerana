package provider

import (
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
)

// Server is the main server struct.
type Server struct {
	config  *Config
	tasks   map[string]*task
	tracker *acomm.Tracker
}

// Provider is an interface to allow a provider to register its tasks with a
// Server.
type Provider interface {
	RegisterTasks(*Server)
}

// NewServer creates and initializes a new Server.
func NewServer(config *Config) (*Server, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	responseSocket := filepath.Join(
		config.SocketDir(),
		"response",
		config.ServiceName()+".sock")
	tracker, err := acomm.NewTracker(responseSocket, nil, nil, config.RequestTimeout())
	if err != nil {
		return nil, err
	}

	return &Server{
		config:  config,
		tasks:   make(map[string]*task),
		tracker: tracker,
	}, nil
}

// Tracker returns the request/response tracker of the Server.
func (s *Server) Tracker() *acomm.Tracker {
	return s.tracker
}

// RegisterTask registers a new task and its handler with the server.
func (s *Server) RegisterTask(taskName string, handler TaskHandler) {
	s.tasks[taskName] = newTask(taskName, s.TaskSocketPath(taskName), s.config.TaskTimeout(taskName), handler)
}

// TaskSocketPath returns the unix socket path for a task
func (s *Server) TaskSocketPath(taskName string) string {
	return filepath.Join(
		s.config.SocketDir(),
		taskName,
		strconv.Itoa(s.config.TaskPriority(taskName))+"-"+s.config.ServiceName()+".sock")
}

// RegisteredTasks returns a list of registered task names.
func (s *Server) RegisteredTasks() []string {
	taskNames := make([]string, 0, len(s.tasks))
	for taskName := range s.tasks {
		taskNames = append(taskNames, taskName)
	}
	return taskNames
}

// Start starts up all of the registered tasks and response handling
func (s *Server) Start() error {
	if err := s.tracker.Start(); err != nil {
		return err
	}

	for _, t := range s.tasks {
		if err := t.start(); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops all of the registered tasks and response handling. Blocks until complete.
func (s *Server) Stop() {
	// Stop all actively handled tasks
	var taskWG sync.WaitGroup
	for _, t := range s.tasks {
		taskWG.Add(1)
		go func(t *task) {
			defer taskWG.Done()
			t.stop()
		}(t)
	}
	taskWG.Wait()

	s.tracker.Stop()
	return
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
