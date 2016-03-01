package acomm_test

import (
	"bytes"
	"io/ioutil"

	"github.com/mistifyio/acomm"
)

func (s *TrackerTestSuite) TestNewStreamUnix() {
	if !s.NoError(s.Tracker.Start(), "failed to start Tracker") {
		return
	}
	data := []byte("foobar")
	src := ioutil.NopCloser(bytes.NewReader(data))

	addr, err := s.Tracker.NewStreamUnix("", nil)
	s.Nil(addr, "shouldn't be able to create stream without src")
	s.Error(err, "shouldn't be able to create stream without src")

	addr, err = s.Tracker.NewStreamUnix("", src)
	s.NotNil(addr, "should create streqm with src")
	s.NoError(err, "should create stream with src")
}

func (s *TrackerTestSuite) TestStreamUnix() {
	if !s.NoError(s.Tracker.Start(), "failed to start Tracker") {
		return
	}
	data := []byte("foobar")
	src := ioutil.NopCloser(bytes.NewReader(data))
	addr, err := s.Tracker.NewStreamUnix("", src)
	if !s.NoError(err) {
		return
	}

	// Unix
	var dest bytes.Buffer
	s.Error(acomm.Stream(nil, addr), "should fail without dest")
	s.Equal(0, dest.Len(), "should not have streamed any data")

	s.Error(acomm.Stream(&dest, nil), "should fail without addr")
	s.Equal(0, dest.Len(), "should not have streamed any data")

	s.NoError(acomm.Stream(&dest, addr), "unix stream should not fail")
	s.Equal(data, dest.Bytes(), "unix stream should have streamed data")
	dest.Reset()

	s.Error(acomm.Stream(&dest, addr), "stream should only be available once")
	s.Equal(0, dest.Len(), "should not have streamed any data")
}

func (s *TrackerTestSuite) TestStreamHTTP() {
	if !s.NoError(s.Tracker.Start(), "failed to start Tracker") {
		return
	}
	data := []byte("foobar")
	src := ioutil.NopCloser(bytes.NewReader(data))
	addr, err := s.Tracker.NewStreamUnix("", src)
	if !s.NoError(err) {
		return
	}
	var dest bytes.Buffer

	// HTTP
	httpAddr, err := s.Tracker.ProxyStreamHTTPURL(nil)
	s.Nil(httpAddr, "shouldn't be able to create HTTP url without stream addr")
	s.Error(err, "shouldn't be able to create HTTP url without stream addr")

	httpAddr, err = s.Tracker.ProxyStreamHTTPURL(addr)
	s.NotNil(httpAddr, "should be able to create HTTP url with stream addr")
	s.NoError(err, "should be able to create HTTP url with stream addr")

	s.NoError(acomm.Stream(&dest, httpAddr), "http stream should not fail")
	s.Equal(data, dest.Bytes(), "http stream should have streamed data")
	dest.Reset()
}
