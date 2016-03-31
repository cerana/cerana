package systemd_test

import (
	"github.com/cerana/cerana/acomm"
	systemdp "github.com/cerana/cerana/providers/systemd"
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
