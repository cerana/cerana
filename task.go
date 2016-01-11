package simple

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/async-comm"
	logx "github.com/mistifyio/mistify-logrus-ext"
)

// TaskHandler if the request handler function for a particular task. It should
// return results or an error, but not both.
type TaskHandler func(interface{}) (interface{}, error)

// Task contains the request listener and handler for a task.
type Task struct {
	Name       string
	Timeout    time.Duration
	SocketPath string
	Listener   *net.UnixListener
	Handler    TaskHandler
	waitgroup  sync.WaitGroup
	stopChan   chan struct{}
}

// newTask creates and initializes a new Task.
func newTask(name, socketPath string, timeout time.Duration, handler TaskHandler) *Task {
	// TODO: Some basic validation
	return &Task{
		Name:       name,
		Timeout:    timeout,
		SocketPath: socketPath,
		Handler:    handler,
		stopChan:   make(chan struct{}),
	}
}

// start starts the task, listening and handling requests.
func (t *Task) start() error {
	if t.Listener != nil {
		return nil
	}

	t.stopChan = make(chan struct{})

	if err := t.createListener(); err != nil {
		return err
	}

	go t.listen()
	return nil
}

// stop shuts down and removes the listener after all requests have been handled.
func (t *Task) stop() {
	// check if listener exists. False->return
	if t.Listener == nil {
		return
	}

	// wait until requests have been serviced
	close(t.stopChan)
	t.waitgroup.Wait()

	t.Listener = nil
	return
}

// createListener creates a unix socket listener.
func (t *Task) createListener() error {
	addr, err := net.ResolveUnixAddr("unix", t.SocketPath)
	if err != nil {
		log.WithFields(log.Fields{
			"task":       t.Name,
			"error":      err,
			"socketPath": t.SocketPath,
		}).Error("failed to resolve response listener unix addr")
		return err
	}

	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		log.WithFields(log.Fields{
			"task":       t.Name,
			"error":      err,
			"socketPath": t.SocketPath,
		}).Error("failed to create response listener")
		return err
	}
	t.Listener = listener

	return nil
}

func (t *Task) listen() {
	// TODO: Defer server wg done?
	defer logx.LogReturnedErr(t.Listener.Close, log.Fields{
		"task":       t.Name,
		"socketPath": t.SocketPath,
	}, "failed to close listener")

	for {
		select {
		case <-t.stopChan:
			log.WithFields(log.Fields{
				"task": t.Name,
			}).Info("stop listening")
			return
		default:
		}

		if err := t.Listener.SetDeadline(time.Now().Add(t.Timeout)); err != nil {
			log.WithFields(log.Fields{
				"task":  t.Name,
				"error": err,
			}).Error("failed to set listener deadline")
		}

		conn, err := t.Listener.Accept()
		if nil != err {
			// Don't worry about a timeout
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}

			log.WithFields(log.Fields{
				"task":  t.Name,
				"error": err,
			}).Error("failed to accept new connection")
			continue
		}

		t.waitgroup.Add(1)
		go t.acceptRequest(conn)
	}
}

func (t *Task) acceptRequest(conn net.Conn) {
	defer t.waitgroup.Done()
	defer logx.LogReturnedErr(conn.Close,
		log.Fields{"task": t.Name},
		"failed to close unix connection",
	)
	if err := conn.SetDeadline(time.Now().Add(t.Timeout)); err != nil {
		log.WithFields(log.Fields{
			"task":  t.Name,
			"error": err,
		}).Error("failed to set connection deadline")
	}

	data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.WithFields(log.Fields{
			"task":  t.Name,
			"error": err,
		}).Error("failed to read request data")
		return
	}

	req := &acomm.Request{}
	if err := json.Unmarshal(data, req); err != nil {
		log.WithFields(log.Fields{
			"task":  t.Name,
			"error": err,
			"data":  string(data),
		}).Error("failed to unmarshal request data")
		return
	}

	// Respond to the initial request
	// TODO: What should the result be? Does it matter as long as err is nil?
	resp, err := acomm.NewResponse(req, struct{}{}, nil)
	respJSON, err := json.Marshal(resp)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"req":      req,
			"response": resp,
		}).Error("failed to marshal initial response")
		return
	}

	if _, err := conn.Write(respJSON); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"req":   req,
			"resp":  resp,
		}).Error("failed to send initial response")
		return
	}

	// Actually perform the task
	t.waitgroup.Add(1)
	go t.handleRequest(req)
}

// handleRequest runs the task-specific handler and sends the results to the
// request's response hook.
func (t *Task) handleRequest(req *acomm.Request) {
	defer t.waitgroup.Done()

	// Run the task-specific request handler
	result, taskErr := t.Handler(req.Args)

	// Note: The acomm calls log the error already, but we want to have a log
	// of the request and response data as well.
	resp, err := acomm.NewResponse(req, result, taskErr)
	if err != nil {
		log.WithFields(log.Fields{
			"task":       t.Name,
			"req":        req,
			"taskResult": result,
			"taskErr":    taskErr,
			"error":      err,
		}).Error("failed to create response")
		return
	}

	if err := req.Respond(resp); err != nil {
		log.WithFields(log.Fields{
			"task":       t.Name,
			"req":        req,
			"taskResult": result,
			"taskErr":    taskErr,
			"error":      err,
		}).Error("failed to send response")
		return
	}
}
