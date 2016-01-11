package acomm_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/async-comm"
	"github.com/stretchr/testify/suite"
)

type TrackerTestSuite struct {
	suite.Suite
	Tracker    *acomm.Tracker
	RespServer *httptest.Server
	Responses  chan *acomm.Response
	Request    *acomm.Request
}

func (s *TrackerTestSuite) SetupSuite() {
	log.SetLevel(log.FatalLevel)
	s.Responses = make(chan *acomm.Response, 10)

	// Mock HTTP response server
	s.RespServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := &acomm.Response{}
		body, err := ioutil.ReadAll(r.Body)
		s.NoError(err, "should not fail reading body")
		s.NoError(json.Unmarshal(body, resp), "should not fail unmarshalling response")
		s.Responses <- resp
	}))
}

func (s *TrackerTestSuite) SetupTest() {
	var err error

	s.Request, err = acomm.NewRequest(s.RespServer.URL, nil, nil, nil)
	s.Require().NoError(err, "request should be valid")

	s.Tracker = acomm.NewTracker("")
	s.Require().NotNil(s.Tracker, "failed to create new Tracker")
}

func (s *TrackerTestSuite) TearDownTest() {
	s.NoError(s.Tracker.StopListener(0 * time.Second))
}

func (s *TrackerTestSuite) TearDownSuite() {
	s.RespServer.Close()
}

func TestTrackerTestSuite(t *testing.T) {
	suite.Run(t, new(TrackerTestSuite))
}

func (s *TrackerTestSuite) NextResp() *acomm.Response {
	timeout := make(chan struct{}, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- struct{}{}
	}()

	var resp *acomm.Response
	select {
	case resp = <-s.Responses:
	case <-timeout:
	}
	return resp
}

func (s *TrackerTestSuite) TestTrackRequest() {
	s.Tracker.TrackRequest(s.Request)
	s.Tracker.TrackRequest(s.Request)
	s.Equal(1, s.Tracker.NumRequests(), "should have one unique request tracked")
}

func (s *TrackerTestSuite) TestRetrieveRequest() {
	response, err := acomm.NewResponse(s.Request, struct{}{}, nil)
	s.NoError(err, "response should be valid")

	s.Nil(s.Tracker.RetrieveRequest(response.ID))
	s.Tracker.TrackRequest(s.Request)
	s.Equal(s.Request, s.Tracker.RetrieveRequest(response.ID), "should retrieve corresponding request")
	s.Equal(0, s.Tracker.NumRequests(), "should have removed tracked request")
}

func (s *TrackerTestSuite) TestStartAndStopListener() {
	stopTimeout := 2 * time.Second

	s.NoError(s.Tracker.StopListener(stopTimeout), "stopping an unstarted should not error")
	s.NoError(s.Tracker.StartListener(), "starting an unstarted should not error")
	s.NoError(s.Tracker.StartListener(), "starting an started should not error")

	s.Tracker.TrackRequest(s.Request)

	s.Equal(errors.New("timeout"), s.Tracker.StopListener(stopTimeout), "stopping a started with requests should error with timeout")
	s.NoError(s.Tracker.StopListener(stopTimeout), "stopping a stopped should not error")
}

func (s *TrackerTestSuite) TestProxyUnix() {
	unixReq, err := s.Tracker.ProxyUnix(s.Request)
	s.Error(err, "should fail to proxy when tracker is not listening")
	s.Nil(unixReq, "should not return a request")

	if !s.NoError(s.Tracker.StartListener(), "listner should start") {
		return
	}

	unixReq, err = s.Tracker.ProxyUnix(s.Request)
	s.NoError(err, "should not fail proxying when tracker is listening")
	s.NotNil(unixReq, "should return a request")
	s.Equal(s.Request.ID, unixReq.ID, "new request should share ID with original")
	s.Equal("unix", unixReq.ResponseHook.Scheme, "new request should have a unix response hook")
	s.Equal(1, s.Tracker.NumRequests(), "should have tracked the new request")

	resp, err := acomm.NewResponse(unixReq, struct{}{}, nil)
	if !s.NoError(err, "new response should not error") {
		return
	}
	if !s.NoError(resp.Send(unixReq.ResponseHook), "response send should not error") {
		return
	}
	time.Sleep(100 * time.Millisecond)
	lastResp := s.NextResp()
	s.Equal(resp.ID, lastResp.ID, "response should have been proxied to original http response hook")
	s.Equal(0, s.Tracker.NumRequests(), "should have removed the request from tracking")

	// Should not proxy a request already using unix response hook
	origUnixReq, err := acomm.NewRequest("unix://foo", struct{}{}, nil, nil)
	if !s.NoError(err, "new request shoudl not error") {
		return
	}
	unixReq, err = s.Tracker.ProxyUnix(origUnixReq)
	s.NoError(err, "should not error with unix response hook")
	s.Equal(origUnixReq, unixReq, "should not proxy unix response hook")
	s.Equal(0, s.Tracker.NumRequests(), "should not response an unproxied request")
}
