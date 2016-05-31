package health

import (
	"errors"
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
)

type Mock struct {
	Data MockData
}

type MockData struct {
	Uptime      bool
	File        bool
	TCPResponse bool
	HTTPStatus  bool
}

func NewMock() *Mock {
	return &Mock{
		Data: MockData{
			Uptime:      true,
			File:        true,
			TCPResponse: true,
			HTTPStatus:  true,
		},
	}
}

// RegisterTasks registers all of the Mock health task handlers with the server.
func (m *Mock) RegisterTasks(server *provider.Server) {
	server.RegisterTask("health-uptime", m.Uptime)
	server.RegisterTask("health-file", m.File)
	server.RegisterTask("health-tcp-response", m.TCPResponse)
	server.RegisterTask("health-http-status", m.HTTPStatus)
}

func (m *Mock) Uptime(req *acomm.Request) (interface{}, *url.URL, error) {
	var err error
	if !m.Data.Uptime {
		err = errors.New("uptime less than expected")
	}
	return nil, nil, err
}

func (m *Mock) File(req *acomm.Request) (interface{}, *url.URL, error) {
	var err error
	if !m.Data.File {
		err = errors.New("file does not exist")
	}
	return nil, nil, err
}

func (m *Mock) TCPResponse(req *acomm.Request) (interface{}, *url.URL, error) {
	var err error
	if !m.Data.TCPResponse {
		err = errors.New("response did not match")
	}
	return nil, nil, err
}

func (m *Mock) HTTPStatus(req *acomm.Request) (interface{}, *url.URL, error) {
	var err error
	if !m.Data.HTTPStatus {
		err = errors.New("unexpected response status code")
	}
	return nil, nil, err
}
