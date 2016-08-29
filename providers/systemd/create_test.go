package systemd_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/cerana/cerana/acomm"
	systemdp "github.com/cerana/cerana/providers/systemd"
	"github.com/coreos/go-systemd/unit"
)

func (s *systemd) TestCreate() {
	tests := []struct {
		name      string
		options   []*unit.UnitOption
		overwrite bool
		err       string
	}{
		{"", nil, false, "missing arg: name"},
		{"empty.service", nil, false, ""},
		{"nonempty.service", []*unit.UnitOption{{"foo", "bar", "baz"}}, false, ""},
		{"nonempty.service", []*unit.UnitOption{{"foo2", "bar2", "baz2"}}, false, "unit file already exists"}, // duplicate
		{"nonempty.service", []*unit.UnitOption{{"foo2", "bar2", "baz2"}}, true, ""},
	}

	for _, test := range tests {
		args := &systemdp.CreateArgs{Name: test.name, UnitOptions: test.options, Overwrite: test.overwrite}
		argsS := fmt.Sprintf("%+v", args)

		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task:         "systemd-create",
			ResponseHook: s.responseHook,
			Args:         args,
		})
		s.Require().NoError(err, argsS)

		res, streamURL, err := s.systemd.Create(req)
		s.Nil(streamURL, argsS)
		s.Nil(res, argsS)
		if test.err == "" {
			if !s.NoError(err, argsS) {
				continue
			}
			p, _ := filepath.Abs(filepath.Join(s.dir, filepath.Base(test.name)))
			contents, err := ioutil.ReadFile(p)
			if !s.NoError(err, argsS) {
				continue
			}
			expected, err := ioutil.ReadAll(unit.Serialize(test.options))
			if !s.NoError(err, argsS) {
				continue
			}
			s.Equal(expected, contents, argsS)
		} else {
			s.EqualError(err, test.err, argsS)
		}
	}
}
