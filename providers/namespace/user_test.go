package namespace_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/providers/namespace"
)

func (s *Namespace) TestSetUser() {
	tests := []struct {
		UIDs        string
		GIDs        string
		expectedErr string
	}{
		{"0 1 1", "0 1 1", ""},
		{"0 1 1\n1 2 1", "0 1 1\n1 2 1", ""},
		{"0 1 0", "0 1 0", os.ErrInvalid.Error()},
		{"0 1 1\n1 2 1\n1 3 1\n1 4 1\n1 5 1\n1 6 1", "0 1 1\n1 2 1\n1 3 1\n1 4 1\n1 5 1\n1 6 1", os.ErrInvalid.Error()},
	}

	for _, test := range tests {
		desc := fmt.Sprintf("%+v", test)
		proc, err := newProc()
		s.Require().NoError(err, desc)
		defer logrusx.LogReturnedErr(proc.Kill, nil, "failed to kill proc")

		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task:         "namespace-set-user",
			ResponseHook: s.responseHook,
			Args: namespace.UserArgs{
				PID:  uint64(proc.Pid),
				UIDs: stringToIDMaps(test.UIDs),
				GIDs: stringToIDMaps(test.GIDs),
			},
		})

		result, streamURL, err := s.namespace.SetUser(req)
		s.Nil(streamURL, desc)
		s.Nil(result, desc)
		if test.expectedErr != "" {
			if !s.NotNil(err, desc) {
				continue
			}
			s.EqualError(err.(*os.PathError).Err, test.expectedErr, desc)
			continue
		} else {
			s.Nil(err, desc)
		}

		s.checkMapFile(proc.Pid, "uid_map", test.UIDs, desc)
		s.checkMapFile(proc.Pid, "gid_map", test.GIDs, desc)
	}
}

func stringToIDMaps(in string) []namespace.IDMap {
	lines := strings.Split(in, "\n")
	maps := make([]namespace.IDMap, len(lines))
	for i, line := range lines {
		parts := strings.Split(line, " ")
		id, _ := strconv.ParseUint(parts[0], 10, 64)
		hostID, _ := strconv.ParseUint(parts[1], 10, 64)
		length, _ := strconv.ParseUint(parts[2], 10, 64)
		maps[i] = namespace.IDMap{
			ID:     id,
			HostID: hostID,
			Length: length,
		}
	}
	return maps
}

func (s *Namespace) checkMapFile(pid int, filename, expected, desc string) bool {
	pass := true
	mapFile, err := os.Open(fmt.Sprintf("/proc/%d/%s", pid, filename))
	if !s.NoError(err, desc) {
		return false
	}
	defer logrusx.LogReturnedErr(mapFile.Close, nil, "failed to close "+filename)

	info, err := mapFile.Stat()
	if !s.NoError(err, desc) {
		return false
	}
	pass = pass && s.Equal(os.FileMode(0644), info.Mode(), desc)

	contents, err := ioutil.ReadAll(mapFile)
	if !s.NoError(err, desc) {
		return false
	}

	// After the file is written, there are whitespace adjustments outside of
	// Namespace's control. Adjust for comparison.
	parts := strings.Split(strings.TrimSpace(string(contents)), "\n")
	r, _ := regexp.Compile(`\s+`)
	for i, s := range parts {
		parts[i] = r.ReplaceAllString(strings.TrimSpace(s), " ")
	}
	pass = pass && s.EqualValues(expected, strings.Join(parts, "\n"), desc)

	return pass
}
