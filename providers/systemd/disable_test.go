package systemd_test

import (
	"fmt"

	"github.com/mistifyio/mistify/acomm"
	systemdp "github.com/mistifyio/mistify/providers/systemd"
)

func (s *systemd) TestDisable() {
	s.Require().NoError(enableAll("./_test"))
	defer func() {
		s.Require().NoError(disableAll("./_test"))
	}()

	tests := []struct {
		name    string
		runtime bool
		err     string
	}{
		{"", false, "missing arg: name"},
		{"doesnotexist.service", false, ""},
		{"systemd-test-loop.service", false, ""},
	}

	for _, test := range tests {
		args := &systemdp.DisableArgs{Name: test.name, Runtime: test.runtime}
		argsS := fmt.Sprintf("%+v", test)

		req, err := acomm.NewRequest("zfs-disable", "unix:///tmp/foobar", "", args, nil, nil)
		s.Require().NoError(err, argsS)

		res, streamURL, err := s.systemd.Disable(req)
		s.Nil(streamURL, argsS)
		s.Nil(res, argsS)
		if test.err == "" {
			s.NoError(err, argsS)
		} else {
			s.EqualError(err, test.err, argsS)
		}
	}
}
