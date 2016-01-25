package simple

import (
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/acomm"
)

// TaskHandler if the request handler function for a particular task. It should
// return results or an error, but not both.
type TaskHandler func(*acomm.Request) (interface{}, error)

// task contains the request listener and handler for a task.
type task struct {
	name        string
	handler     TaskHandler
	reqTimeout  time.Duration
	reqListener *acomm.UnixListener
	waitgroup   sync.WaitGroup
}

// newTask creates and initializes a new task.
func newTask(name, socketPath string, reqTimeout time.Duration, handler TaskHandler) *task {
	return &task{
		name:        name,
		handler:     handler,
		reqTimeout:  reqTimeout,
		reqListener: acomm.NewUnixListener(socketPath),
	}
}

// start starts the task handler.
func (t *task) start() error {
	if err := t.reqListener.Start(); err != nil {
		return err
	}

	go t.handleConns()
	return nil
}

// stop shuts down the task handler.
func (t *task) stop() {
	// Stop request listener and handle all open connections
	t.reqListener.Stop()

	// Wait for all actively handled requests
	t.waitgroup.Wait()
}

func (t *task) handleConns() {
	for {
		conn := t.reqListener.NextConn()
		if conn == nil {
			return
		}
		go t.acceptRequest(conn)
	}
}

func (t *task) acceptRequest(conn net.Conn) {
	defer t.reqListener.DoneConn(conn)
	var respErr error

	req := &acomm.Request{}
	if err := acomm.UnmarshalConnData(conn, req); err != nil {
		respErr = err
	}

	if err := req.Validate(); err != nil {
		respErr = err
	}

	// Respond to the initial request
	resp, err := acomm.NewResponse(req, nil, respErr)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"req":     req,
			"respErr": respErr,
		}).Error("failed to create initial response")
		return
	}

	if acomm.SendConnData(conn, resp); err != nil {
		return
	}

	if respErr != nil {
		return
	}
	// Actually perform the task
	t.waitgroup.Add(1)
	go t.handleRequest(req)
}

// handleRequest runs the task-specific handler and sends the results to the
// request's response hook.
func (t *task) handleRequest(req *acomm.Request) {
	defer t.waitgroup.Done()

	// Run the task-specific request handler
	result, taskErr := t.handler(req)

	// Note: The acomm calls log the error already, but we want to have a log
	// of the request and response data as well.
	resp, err := acomm.NewResponse(req, result, taskErr)
	if err != nil {
		log.WithFields(log.Fields{
			"task":       t.name,
			"req":        req,
			"taskResult": result,
			"taskErr":    taskErr,
			"error":      err,
		}).Error("failed to create response")
		return
	}

	if err := req.Respond(resp); err != nil {
		log.WithFields(log.Fields{
			"task":       t.name,
			"req":        req,
			"taskResult": result,
			"taskErr":    taskErr,
			"error":      err,
		}).Error("failed to send response")
		return
	}
}
