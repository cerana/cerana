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
		name    string
		options []*unit.UnitOption
		err     string
	}{
		{"", nil, "missing arg: name"},
		{"empty.service", nil, ""},
		{"nonempty.service", []*unit.UnitOption{{"foo", "bar", "baz"}}, ""},
		{"nonempty.service", []*unit.UnitOption{{"foo2", "bar2", "baz2"}}, "unit file already exists"}, // duplicate
	}

	for _, test := range tests {
		args := &systemdp.CreateArgs{Name: test.name, UnitOptions: test.options}
		argsS := fmt.Sprintf("%+v", args)

		req := acomm.NewRequest("systemd-create")
		req.ResponseHook = s.responseHook
		s.Require().NoError(req.SetArgs(args), argsS)

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
