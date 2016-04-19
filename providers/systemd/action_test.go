package systemd_test

import (
	"fmt"
	"net/url"

	"github.com/cerana/cerana/acomm"
	systemdp "github.com/cerana/cerana/providers/systemd"
)

func (s *systemd) TestActions() {
	s.Require().NoError(enableAll("./_test_data"))
	defer func() {
		s.Require().NoError(disableAll("./_test_data"))
	}()

	tests := []struct {
		action string
		name   string
		mode   string
		err    string
	}{
		{"start", "", "", "missing arg: name"},
		{"start", "doesnotexist.service", systemdp.ModeFail, "unit not found"},
		{"start", "dbus.service", systemdp.ModeFail, ""},
		{"start", "systemd-test-exit0.service", systemdp.ModeFail, ""},
		{"start", "systemd-test-exit1.service", systemdp.ModeFail, ""},
		{"start", "systemd-test-loop.service", systemdp.ModeFail, ""},

		{"restart", "systemd-test-exit0.service", systemdp.ModeFail, ""},
		{"restart", "systemd-test-loop.service", systemdp.ModeFail, ""},

		{"stop", "systemd-test-loop.service", systemdp.ModeFail, ""},
		{"stop", "systemd-test-exit0.service", systemdp.ModeFail, ""},
	}

	for _, test := range tests {
		args := &systemdp.ActionArgs{Name: test.name, Mode: test.mode}
		argsS := fmt.Sprintf("%+v", test)

		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task:         "systemd-" + test.action,
			ResponseHook: s.responseHook,
			Args:         args,
		})
		s.Require().NoError(err, argsS)

		var fn func(*acomm.Request) (interface{}, *url.URL, error)
		switch test.action {
		case "start":
			fn = s.systemd.Start
		case "stop":
			fn = s.systemd.Stop
		case "restart":
			fn = s.systemd.Restart
		}
		res, streamURL, err := fn(req)
		s.Nil(streamURL, argsS)
		s.Nil(res, argsS)
		if test.err == "" {
			s.NoError(err, argsS)
		} else {
			s.EqualError(err, test.err, argsS)
		}
	}
}
