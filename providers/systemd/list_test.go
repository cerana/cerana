package systemd_test

import (
	"github.com/mistifyio/mistify/acomm"
	"github.com/mistifyio/mistify/providers/systemd"
)

func (s *sd) TestList() {
	res, streamURL, err := s.systemd.List(&acomm.Request{})
	s.Nil(streamURL)
	if !s.NoError(err) {
		return
	}
	result, ok := res.(*systemd.ListResult)
	if !s.True(ok) {
		return
	}
	s.True(len(result.Units) > 0)
}
