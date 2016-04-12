package systemd_test

import (
	"fmt"
	"os"

	"github.com/cerana/cerana/acomm"
	systemdp "github.com/cerana/cerana/providers/systemd"
)

func (s *systemd) TestRemove() {
	tests := []struct {
		name  string
		exist bool
		err   string
	}{
		{"", false, "missing arg: name"},
		{"foo", false, fmt.Sprintf("remove %s: no such file or directory", s.dir+"/foo")},
		{"foo", true, ""},
		{".", false, "invalid name"},
	}

	for _, test := range tests {
		args := &systemdp.RemoveArgs{Name: test.name}
		argsS := fmt.Sprintf("%+v", args)

		unitpath := s.dir + "/" + test.name

		if test.exist {
			f, err := os.Create(unitpath)
			s.Require().NoError(err)
			s.NoError(f.Close())
		}

		req := acomm.NewRequest("systemd-remove")
		req.ResponseHook = s.responseHook
		s.Require().NoError(req.SetArgs(args), argsS)

		res, streamURL, err := s.systemd.Remove(req)
		s.Nil(streamURL, argsS)
		s.Nil(res, argsS)
		if test.err == "" {
			if !s.NoError(err, argsS) {
				continue
			}
			_, err = os.Stat(unitpath)
			s.True(os.IsNotExist(err), argsS)
		} else {
			s.EqualError(err, test.err, argsS)
		}
	}
}
