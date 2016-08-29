package service

import (
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
)

type continueCheck func(*acomm.Response) (bool, error)

// Provider is a provider of service management functionality.
type Provider struct {
	config  *Config
	tracker *acomm.Tracker
}

// New creates a new instance of Provider.
func New(config *Config, tracker *acomm.Tracker) *Provider {
	return &Provider{
		config:  config,
		tracker: tracker,
	}
}

// RegisterTasks registers all of the provider task handlers with the server.
func (p *Provider) RegisterTasks(server *provider.Server) {
	server.RegisterTask("service-create", p.Create)
	server.RegisterTask("service-get", p.Get)
	server.RegisterTask("service-list", p.List)
	server.RegisterTask("service-restart", p.Restart)
	server.RegisterTask("service-remove", p.Remove)
}

func (p *Provider) executeRequests(requests []*acomm.Request, continueChecks []continueCheck) error {
	if continueChecks == nil {
		continueChecks = make([]continueCheck, len(requests))
	}
	for i, req := range requests {
		doneChan := make(chan *acomm.Response, 1)
		defer close(doneChan)
		rh := func(req *acomm.Request, resp *acomm.Response) {
			doneChan <- resp
		}
		req.SuccessHandler = rh
		req.ErrorHandler = rh

		if err := p.tracker.TrackRequest(req, p.config.RequestTimeout()); err != nil {
			return err
		}
		if err := acomm.Send(p.config.CoordinatorURL(), req); err != nil {
			return err
		}

		resp := <-doneChan
		if resp.Error != nil {
			return resp.Error
		}

		if check := continueChecks[i]; check != nil {
			next, err := check(resp)
			if err != nil {
				return err
			}
			if !next {
				return nil
			}
		}
	}
	return nil
}
