package systemd_test

import (
	"fmt"

	"github.com/mistifyio/mistify/acomm"
	systemdp "github.com/mistifyio/mistify/providers/systemd"
)

func (s *systemd) TestEnable() {
	defer func() {
		s.Require().NoError(disableAll("./_test"))
	}()

	tests := []struct {
		name    string
		runtime bool
		force   bool
		err     string
	}{
		{"", false, false, "missing arg: name"},
		{"doesnotexist.service", false, false, "No such file or directory"},
		{"systemd-test-loop.service", false, false, ""},
	}

	for _, test := range tests {
		args := &systemdp.EnableArgs{test.name, test.runtime, test.force}
		argsS := fmt.Sprintf("%+v", test)

		req, err := acomm.NewRequest("zfs-enable", "unix:///tmp/foobar", "", args, nil, nil)
		s.Require().NoError(err, argsS)

		res, streamURL, err := s.systemd.Enable(req)
		s.Nil(streamURL, argsS)
		s.Nil(res, argsS)
		if test.err == "" {
			s.NoError(err, argsS)
		} else {
			s.EqualError(err, test.err, argsS)
		}
	}
}
