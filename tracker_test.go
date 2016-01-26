package acomm_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/acomm"
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

	s.Request, err = acomm.NewRequest("foobar", s.RespServer.URL, nil, nil, nil)
	s.Require().NoError(err, "request should be valid")

	s.Tracker, err = acomm.NewTracker("", "")
	s.Require().NoError(err, "failed to create new Tracker")
	s.Require().NotNil(s.Tracker, "failed to create new Tracker")
}

func (s *TrackerTestSuite) TearDownTest() {
	s.Tracker.Stop()
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
		time.Sleep(5 * time.Second)
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
	if !s.NoError(s.Tracker.Start(), "should have started tracker") {
		return
	}
	s.NoError(s.Tracker.TrackRequest(s.Request), "should have successfully tracked request")
	s.Error(s.Tracker.TrackRequest(s.Request), "duplicate ID should have failed to track request")
	s.Equal(1, s.Tracker.NumRequests(), "should have one unique request tracked")
	s.True(s.Tracker.RemoveRequest(s.Request))
}

func (s *TrackerTestSuite) TestStartAndStopListener() {
	s.Tracker.Stop()
	s.NoError(s.Tracker.Start(), "starting an unstarted should not error")
	s.NoError(s.Tracker.Start(), "starting an started should not error")

	s.NoError(s.Tracker.TrackRequest(s.Request), "should have successfully tracked request")

	go s.Tracker.RemoveRequest(s.Request)

	s.Tracker.Stop()
}

func (s *TrackerTestSuite) TestProxyUnix() {
	unixReq, err := s.Tracker.ProxyUnix(s.Request)
	s.Error(err, "should fail to proxy when tracker is not listening")
	s.Nil(unixReq, "should not return a request")

	if !s.NoError(s.Tracker.Start(), "listner should start") {
		return
	}

	unixReq, err = s.Tracker.ProxyUnix(s.Request)
	s.NoError(err, "should not fail proxying when tracker is listening")
	s.NotNil(unixReq, "should return a request")
	s.Equal(s.Request.ID, unixReq.ID, "new request should share ID with original")
	s.Equal("unix", unixReq.ResponseHook.Scheme, "new request should have a unix response hook")
	s.Equal(1, s.Tracker.NumRequests(), "should have tracked the new request")
	resp, err := acomm.NewResponse(unixReq, struct{}{}, nil, nil)
	if !s.NoError(err, "new response should not error") {
		return
	}
	if !s.NoError(acomm.Send(unixReq.ResponseHook, resp), "response send should not error") {
		return
	}

	lastResp := s.NextResp()
	if !s.NotNil(lastResp, "response should have been proxied to original http response hook") {
		return
	}
	s.Equal(resp.ID, lastResp.ID, "response should have been proxied to original http response hook")
	s.Equal(0, s.Tracker.NumRequests(), "should have removed the request from tracking")

	// Should not proxy a request already using unix response hook
	origUnixReq, err := acomm.NewRequest("foobar", "unix://foo", struct{}{}, nil, nil)
	if !s.NoError(err, "new request shoudl not error") {
		return
	}
	unixReq, err = s.Tracker.ProxyUnix(origUnixReq)
	s.NoError(err, "should not error with unix response hook")
	s.Equal(origUnixReq, unixReq, "should not proxy unix response hook")
	s.Equal(0, s.Tracker.NumRequests(), "should not response an unproxied request")
}
