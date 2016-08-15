package provider

import (
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
)

// TaskHandler if the request handler function for a particular task. It should
// return results or an error, but not both.
type TaskHandler func(*acomm.Request) (interface{}, *url.URL, error)

// task contains the request listener and handler for a task.
type task struct {
	name         string
	providerName string
	handler      TaskHandler
	reqTimeout   time.Duration
	reqListener  *acomm.UnixListener
	waitgroup    sync.WaitGroup
}

// newTask creates and initializes a new task.
func newTask(name, providerName, socketPath string, reqTimeout time.Duration, handler TaskHandler) *task {
	return &task{
		name:         name,
		providerName: providerName,
		handler:      handler,
		reqTimeout:   reqTimeout,
		reqListener:  acomm.NewUnixListener(socketPath, 0),
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
	t.reqListener.Stop(0)

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
		respErr = errors.Wrap(err, "failed to unmarshal request")
	}

	if err := req.Validate(); err != nil {
		respErr = errors.Wrapv(err, map[string]interface{}{"request": req})
	}

	// Respond to the initial request
	resp, err := acomm.NewResponse(req, nil, nil, respErr)
	if err != nil {
		err = errors.Wrapv(err, map[string]interface{}{"request": req, "respErr": respErr})
		logrus.WithField("error", err).Error("failed to create initial response")
		return
	}

	if err := acomm.SendConnData(conn, resp); err != nil {
		logrus.WithField("error", err).Error("failed to send initial response")
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
	result, streamAddr, taskErr := t.handler(req)
	taskErr = errors.Wrap(taskErr, t.providerName, t.Name)
	errData := map[string]interface{}{
		"task":       t.name,
		"request":    req,
		"taskResult": result,
		"streamAddr": streamAddr,
		"taskErr":    taskErr,
	}

	if taskErr != nil {
		err := errors.Wrapv(taskErr, errData)
		logrus.WithField("error", err).Error("task handler error")
	}

	// Note: The acomm calls log the error already, but we want to have a log
	// of the request and response data as well.
	resp, err := acomm.NewResponse(req, result, streamAddr, taskErr)
	if err != nil {
		err = errors.Wrapv(err, errData)
		logrus.WithField("error", err).Error("failed to create response")
		return
	}

	if err := req.Respond(resp); err != nil {
		err = errors.Wrapv(err, errData)
		logrus.WithField("error", err).Error("failed to send response")
		return
	}
}
