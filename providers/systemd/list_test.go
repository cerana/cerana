package systemd_test

import (
	"github.com/mistifyio/mistify/acomm"
	systemdp "github.com/mistifyio/mistify/providers/systemd"
)

func (s *systemd) TestList() {
	res, streamURL, err := s.systemd.List(&acomm.Request{})
	s.Nil(streamURL)
	if !s.NoError(err) {
		return
	}
	result, ok := res.(*systemdp.ListResult)
	if !s.True(ok) {
		return
	}
	s.True(len(result.Units) > 0)
}
