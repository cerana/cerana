package simple

import (
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

// Server is the main server struct.
type Server struct {
	config *Config
	tasks  map[string]*Task
}

// NewServer creates and initializes a new Server.
func NewServer(config *Config) *Server {
	return &Server{
		config: config,
		tasks:  make(map[string]*Task),
	}
}

// RegisterTask registers a new task and its handler with the server.
func (s *Server) RegisterTask(taskName string, handler TaskHandler) {
	socketPath := filepath.Join(
		s.config.SocketDir(),
		taskName,
		strconv.Itoa(s.config.TaskPriority(taskName)),
		s.config.ServiceName())

	task := newTask(taskName, socketPath, s.config.TaskTimeout(taskName), handler)
	s.tasks[taskName] = task
}

// RegisteredTasks returns a list of registered task names.
func (s *Server) RegisteredTasks() []string {
	taskNames := make([]string, 0, len(s.tasks))
	for taskName := range s.tasks {
		taskNames = append(taskNames, taskName)
	}
	return taskNames
}

// Start starts up all of the registered tasks.
func (s *Server) Start() error {
	for _, task := range s.tasks {
		if err := task.start(); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops all of the registered tasks.
func (s *Server) Stop() {
	var wg sync.WaitGroup
	for _, task := range s.tasks {
		wg.Add(1)
		go func(t *Task) {
			defer wg.Done()
			t.stop()
		}(task)
	}
	wg.Wait()
	return
}

// StopOnSignal will wait until one of the specified signals is received and
// then stop the server. If no signals are specified, it will use a default
// set.
func (s *Server) StopOnSignal(signals ...os.Signal) {
	if len(signals) == 0 {
		signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, signals...)
	sig := <-sigChan
	log.WithFields(log.Fields{
		"signal": sig,
	}).Info("signal recieved, stopping")

	s.Stop()
}
