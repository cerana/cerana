package systemd_test

import (
	"github.com/mistifyio/mistify/acomm"
	"github.com/mistifyio/mistify/providers/systemd"
)

func (s *sd) TestListUnits() {
	res, streamURL, err := s.systemd.ListUnits(&acomm.Request{})
	s.Nil(streamURL)
	if !s.NoError(err) {
		return
	}
	result, ok := res.(*systemd.ListUnitsResult)
	if !s.True(ok) {
		return
	}
	s.True(len(result.Units) > 0)
}
