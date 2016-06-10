package health_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/cerana/cerana/acomm"
	healthp "github.com/cerana/cerana/providers/health"
)

func (s *health) TestFile() {
	origUmask := syscall.Umask(0)
	defer syscall.Umask(origUmask)

	dir, err := ioutil.TempDir("", "health")
	s.Require().NoError(err)
	defer func() { _ = os.RemoveAll(dir) }()

	tmp := filepath.Join(dir, "healthFile")
	size := int64(2024)
	mode := os.FileMode(0666)
	s.Require().NoError(ioutil.WriteFile(tmp, make([]byte, size), mode))

	tests := []struct {
		path        string
		notExist    bool
		mode        os.FileMode
		minSize     int64
		maxSize     int64
		expectedErr string
	}{
		{"", false, 0, 0, 0, "missing arg: path"},
		{"/asdf/asdf", false, 0, 0, 0, "stat /asdf/asdf: no such file or directory"},
		{"/asdf/asdf", true, 0, 0, 0, ""},
		{tmp, true, 0, 0, 0, "file exists"},
		{tmp, false, os.FileMode(0777), 0, 0, "unexpected mode"},
		{tmp, false, mode, 0, 0, ""},
		{tmp, false, 0, 0, 0, ""},
		{tmp, false, 0, size + 1, 0, "size below min"},
		{tmp, false, 0, size, 0, ""},
		{tmp, false, 0, size - 1, 0, ""},
		{tmp, false, 0, 0, size - 1, "size above max"},
		{tmp, false, 0, 0, size, ""},
		{tmp, false, 0, 0, size + 1, ""},
	}

	for _, test := range tests {
		desc := fmt.Sprintf("%+v", test)
		args := &healthp.FileArgs{
			Path:     test.path,
			NotExist: test.notExist,
			Mode:     test.mode,
			MinSize:  test.minSize,
			MaxSize:  test.maxSize,
		}

		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task:         "health-file",
			ResponseHook: s.responseHook,
			Args:         args,
		})
		s.Require().NoError(err, desc)

		resp, stream, err := s.health.File(req)
		s.Nil(resp, desc)
		s.Nil(stream, desc)
		if test.expectedErr == "" {
			s.Nil(err, desc)
		} else {
			s.EqualError(err, test.expectedErr, desc)
		}
	}
}
