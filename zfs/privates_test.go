package zfs

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/suite"
)

var (
	eagain       = syscall.EAGAIN.Error()
	ebadf        = syscall.EBADF.Error()
	ebusy        = syscall.EBUSY.Error()
	eexist       = syscall.EEXIST.Error()
	einval       = syscall.EINVAL.Error()
	enametoolong = syscall.ENAMETOOLONG.Error()
	enoent       = syscall.ENOENT.Error()
	epipe        = syscall.EPIPE.Error()
	exdev        = syscall.EXDEV.Error()
)

const (
	// 257 * "z"
	longName = "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"
)

type internal struct {
	pool  string
	files []string
	suite.Suite
}

func TestSuiteInternal(t *testing.T) {
	suite.Run(t, &internal{})
}

func command(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	return cmd
}

func (s *internal) create(pool string) {
	s.pool = pool
	files := make([]string, 5)
	for i := range files {
		f, err := ioutil.TempFile("", "zfs-test-temp")
		if err != nil {
			panic(err)
		}
		files[i] = f.Name()
		s.Require().NoError(f.Close())
	}
	s.files = files

	script := []byte(`
	set -e
	pool=` + s.pool + `
	zpool list $pool &>/dev/null && zpool destroy $pool
	files=(` + strings.Join(files, " ") + `)
	for f in ${files[*]}; do
		truncate -s1G $f
	done
	zpool create $pool ${files[*]}

	zfs create $pool/a
	zfs create $pool/a/1
	zfs create $pool/a/2
	zfs create $pool/a/4
	zfs snapshot $pool/a/1@snap1
	zfs snapshot $pool/a/2@snap1
	zfs snapshot $pool/a/2@snap2
	zfs snapshot $pool/a/2@snap3
	zfs clone $pool/a/1@snap1 $pool/a/3
	zfs hold hold1 $pool/a/2@snap1
	zfs hold hold2 $pool/a/2@snap2
	zfs unmount $pool/a/4

	zfs create $pool/b
	zfs create -V 8192 $pool/b/1
	zfs create -b 1024 -V 2048 $pool/b/2
	zfs create -V 8192 $pool/b/4
	zfs snapshot $pool/b/1@snap1
	zfs snapshot $pool/b/2@snap1
	zfs snapshot $pool/b/2@snap2
	zfs snapshot $pool/b/2@snap3
	zfs clone $pool/b/1@snap1 $pool/b/3
	zfs hold hold1 $pool/b/2@snap1
	zfs hold hold2 $pool/b/2@snap2

	zfs create $pool/c
	zfs create $pool/c/one
	zfs create $pool/c/two
	zfs create $pool/c/three
	zfs snapshot -r $pool/c@snap1

	exit 0
	`)

	cmd := command("sudo", "bash", "-c", string(script))

	stdin, err := cmd.StdinPipe()
	s.Require().NoError(err)
	go func() {
		_, err := stdin.Write([]byte(script))
		s.Require().NoError(err)
	}()

	s.Require().NoError(cmd.Run())
}

func (s *internal) destroy() {
	err := command("sudo", "zpool", "destroy", s.pool).Run()
	for i := range s.files {
		s.NoError(os.Remove(s.files[i]))
	}
	s.Require().NoError(err)
}

func (s *internal) SetupTest() {
	s.create("zfs-test")
}

func (s *internal) TearDownTest() {
	s.destroy()
}

func (s *internal) TestClone() {
	s.EqualError(clone(s.pool+"/a/2", s.pool+"/a/1", nil), eexist)
	s.EqualError(clone(s.pool+"/a 3", s.pool+"/a/1", nil), einval)
	s.EqualError(clone(s.pool+"/a/"+longName, s.pool+"/a/1", nil), einval) // WANTE(ENAMETOOLONG)
	s.EqualError(clone(s.pool+"/a/z", s.pool+"/a/"+longName, nil), enametoolong)
	s.NoError(clone(s.pool+"/a/z", s.pool+"/a/1@snap1", nil))
}

func (s *internal) TestCreateFS() {
	s.EqualError(create("zfs-test-no-exists/1", dmuZFS, nil), enoent)
	s.EqualError(create(s.pool+"/~1", dmuZFS, nil), einval)
	s.EqualError(create(s.pool+"/1", dmuNumtypes+1, nil), einval)
	s.EqualError(create(s.pool+"/1", dmuZFS, map[string]interface{}{"bad-prop": true}), einval)
	s.EqualError(create(s.pool+"/"+longName, dmuZFS, nil), einval) // WANTE(ENAMETOOLONG)
	s.NoError(create(s.pool+"/1", dmuZFS, nil))
	s.EqualError(create(s.pool+"/1", dmuZFS, nil), eexist)
	s.EqualError(create(s.pool+"/2/2", dmuZFS, nil), enoent)
}

func (s *internal) TestCreateVOL() {
	type p map[string]interface{}
	u0 := uint64(0)
	u1 := uint64(1)
	u1024 := uint64(1024)
	u8192 := u1024 * 8

	tests := []struct {
		name  string
		props p
		err   string
	}{
		{"zfs-test-no-exists/1", nil, enoent},
		{s.pool + "/" + longName, nil, einval}, // WANTE(ENAMETOOLONG)
		{s.pool + "/~1", nil, einval},
		{s.pool + "/1", p{"bad-prop": true}, einval},
		{s.pool + "/6", p{"volsize": u0}, einval},
		{s.pool + "/7", p{"volsize": u1}, einval},
		{s.pool + "/8", p{"volsize": u8192 + 1}, einval},
		{s.pool + "/9", p{"volsize": u8192}, ""},
		{s.pool + "/9", p{"volsize": u8192}, eexist},
		{s.pool + "/22", p{"volsize": u0, "volblocksize": u1024}, einval},
		{s.pool + "/23", p{"volsize": u1, "volblocksize": u1024}, einval},
		{s.pool + "/24", p{"volsize": u1024 + 1, "volblocksize": u1024}, einval},
		{s.pool + "/25", p{"volsize": u1024, "volblocksize": u1024}, ""},
		{s.pool + "/25", p{"volsize": u1024, "volblocksize": u1024}, eexist},
		{s.pool + "/26/1", p{"volsize": u1024, "volblocksize": u1024}, enoent},
	}
	for _, test := range tests {
		err := create(test.name, dmuZVOL, test.props)
		if test.err == "" {
			s.NoError(err, "name: %v, props: %v",
				test.name, test.props)
		} else {
			s.EqualError(err, test.err, "name: %v, props: %v",
				test.name, test.props)
		}
	}

}

func (s *internal) unmount(ds string) error {
	return command("sudo", "zfs", "unmount", ds).Run()
}

func (s *internal) unhold(tag, snapshot string) error {
	return command("sudo", "zfs", "release", tag, snapshot).Run()
}

func destroyKids(ds string) error {
	script := `
		ds=` + ds + `
		for fs in $(zfs list -H -t all -r $ds | sort -r | grep -v -E "^$ds[[:space:]]" | awk '{print $1}'); do
			zfs destroy -r $fs
		done
		`
	return command("sudo", "bash", "-c", script).Run()
}

func (s *internal) TestDestroy() {
	s.EqualError(destroy("non-existent-pool", false), enoent)
	s.EqualError(destroy("non-existent-pool", true), enoent)
	s.EqualError(destroy(s.pool+"/"+longName, false), einval) // WANTE(ENAMETOOLONG)
	s.EqualError(destroy(s.pool+"/"+longName, true), einval)  // WANTE(ENAMETOOLONG)
	s.EqualError(destroy(s.pool+"/z", false), enoent)
	s.EqualError(destroy(s.pool+"/z", true), enoent)

	// mounted
	s.EqualError(destroy(s.pool+"/a/3", false), ebusy)
	s.NoError(s.unmount(s.pool + "/a/3"))
	s.NoError(destroy(s.pool+"/a/3", true))

	// has holds
	s.EqualError(destroy(s.pool+"/a/2@snap1", false), ebusy)
	s.NoError(s.unhold("hold1", s.pool+"/a/2@snap1"))
	s.NoError(destroy(s.pool+"/a/2@snap1", false))
	s.EqualError(destroy(s.pool+"/a/2@snap2", false), ebusy)
	s.NoError(s.unhold("hold2", s.pool+"/a/2@snap2"))
	s.NoError(destroy(s.pool+"/a/2@snap2", true))

	s.EqualError(destroy(s.pool+"/b/2@snap1", false), ebusy)
	s.NoError(s.unhold("hold1", s.pool+"/b/2@snap1"))
	s.NoError(destroy(s.pool+"/b/2@snap1", true))
	s.EqualError(destroy(s.pool+"/b/2@snap2", false), ebusy)
	s.NoError(s.unhold("hold2", s.pool+"/b/2@snap2"))
	s.NoError(destroy(s.pool+"/b/2@snap2", true))

	// misc
	s.NoError(destroy(s.pool+"/a/4", true))
	s.EqualError(destroy(s.pool+"/a/4", false), enoent)
	s.EqualError(destroy(s.pool+"/a/4", true), enoent)
	s.NoError(destroy(s.pool+"/b/4", false))
	s.NoError(destroy(s.pool+"/b/3", true))

	// has child datasets
	s.EqualError(destroy(s.pool+"/a", false), ebusy)
	s.NoError(s.unmount(s.pool + "/a"))
	s.EqualError(destroy(s.pool+"/a", false), eexist)
	s.NoError(destroyKids(s.pool + "/a"))
	s.NoError(destroy(s.pool+"/a", false))

	s.EqualError(destroy(s.pool+"/b", false), ebusy)
	s.NoError(destroyKids(s.pool + "/b"))
	s.NoError(s.unmount(s.pool + "/b"))
	s.NoError(destroy(s.pool+"/b", true))

}

func (s *internal) TestExists() {
	s.EqualError(exists("should-not-exist"), enoent)
	s.NoError(exists(s.pool + "/a"))
}

func (s *internal) TestHolds() {
	h, err := holds(s.pool + "should-not-exist")
	s.EqualError(err, enoent)

	h, err = holds(s.pool + "/" + longName)
	s.EqualError(err, einval) // WANTE(ENAMETOOLONG)

	h, err = holds(s.pool + "/a/2@snap42")
	s.EqualError(err, enoent)

	h, err = holds(s.pool + "/a")
	s.NoError(err)
	s.Len(h, 0)

	h, err = holds(s.pool + "/a/2@snap2")
	s.NoError(err)
	s.Len(h, 1)
}

func (s *internal) TestListEmpty() {
	s.destroy()

	m, err := list("", nil, true, 0)
	s.NoError(err, "m: %v", m)

	s.create(s.pool)
}

func (s *internal) TestList() {
	type t map[string]bool
	tests := []struct {
		name    string
		types   t
		recurse bool
		depth   uint64
		expNum  int
		err     string
	}{
		{name: "blah", err: enoent},
		{name: s.pool + "/" + longName, err: einval}, // WANTE(ENAMETOOLONG)
		{name: s.pool, expNum: 1},
		{name: s.pool, recurse: true, expNum: 11},
		{name: s.pool, recurse: true, depth: 1, expNum: 4},
		{name: s.pool + "/a", recurse: true, expNum: 5},
		{name: s.pool + "/a", recurse: true, depth: 1, expNum: 5},
		{name: s.pool, types: t{"volume": true}, recurse: true, expNum: 4},
		{name: s.pool, types: t{"volume": true}, recurse: true, depth: 1, expNum: 0},
		{name: s.pool + "/a", types: t{"volume": true}, recurse: true, expNum: 0},
		{name: s.pool + "/b", types: t{"volume": true}, recurse: true, expNum: 4},
		{name: s.pool + "/b", types: t{"volume": true}, recurse: true, depth: 1, expNum: 4},
		{name: s.pool, types: t{"snapshot": true}, recurse: true, expNum: 12},
		{name: s.pool, types: t{"snapshot": true}, recurse: true, depth: 1, expNum: 0},
		{name: s.pool + "/a", types: t{"snapshot": true}, recurse: true, expNum: 4},
		{name: s.pool + "/a/1", types: t{"snapshot": true}, recurse: true, expNum: 1},
		{name: s.pool + "/b", types: t{"snapshot": true}, recurse: true, expNum: 4},
		{name: s.pool + "/b/1", types: t{"snapshot": true}, recurse: true, expNum: 1},
	}
	for i, test := range tests {
		m, err := list(test.name, test.types, test.recurse, test.depth)
		if test.err != "" {
			s.EqualError(err, test.err, "test num:%d", i)
			s.Nil(m, "test:%d", i)
		} else {
			s.NoError(err, "test:%d", i)
			s.Len(m, test.expNum, "test num:%d", i)
		}
	}
}

func (s *internal) TestRename() {
	var tests = []struct {
		old   string
		new   string
		err   string
		erred string // implies recursive
	}{
		{"should-not-exist", s.pool + "/blah", enoent, ""},
		{s.pool + "/" + longName, s.pool + "/a/zz", einval, ""}, // WANTE(ENAMETOOLONG)
		{s.pool + "/a/3", s.pool + "/" + longName, einval, ""},  // WANTE(ENAMETOOLONG)
		{s.pool + "/a/3", s.pool + "-not", exdev, ""},
		{s.pool + "/a/2@snap1", s.pool + "/a/2@snap1.1", "", ""},
		//{s.pool + "/c@snap1", @snap2", "", ""},
	}
	for _, test := range tests {
		s.T().Log("old:", test.old, "new:", test.new)

		recursive := test.erred != ""
		if test.err == "" {
			s.NoError(exists(test.old))
			s.EqualError(exists(test.new), enoent)
		}
		erred, err := rename(test.old, test.new, recursive)
		if test.err == "" {
			s.NoError(err)
			s.NoError(exists(test.new))
			s.EqualError(exists(test.old), enoent)
		} else {
			s.EqualError(err, test.err)
			s.Equal(test.erred, erred)
		}
	}
}

func touch(name string) error {
	return command("sudo", "touch", name).Run()
}

func snap(name string) error {
	return command("zfs", "snapshot", name).Run()
}

func (s *internal) TestRollback() {
	var tests = []struct {
		name   string
		latest string
		err    string
	}{
		{"should-not-exist", "", enoent},
		{s.pool + "/" + longName + "@snap3", "", einval}, // WANTE(ENAMETOOLONG)
		{s.pool + "/a/2@snap3", "", einval},
		{"/a/2@snap2", "", einval},
		{"/a/2@snap3", s.pool + "/a/2@snap3", einval},
		{s.pool + "/a/2", s.pool + "/a/2@tempsnap", ""},
	}
	for _, test := range tests {
		s.T().Log("name:", test.name)
		name := "/" + test.name + "/" + "markerfile"
		if test.err == "" {
			NoError := s.Require().NoError
			NoError(touch(name))
			NoError(snap(test.name + "@tempsnap"))
			NoError(command("sudo", "rm", name).Run())
		}
		latest, err := rollback(test.name)
		if test.err == "" {
			s.NoError(err)
			s.Equal(test.latest, latest)
			_, err = os.Stat(name)
			s.NoError(err)
		} else {
			s.EqualError(err, test.err)
		}
	}
}

func (s *internal) receive(r *os.File, target string) error {
	cmd := command("sudo", "zfs", "receive", "-e", target)
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		_, err := io.Copy(in, r)
		if err != nil {
			panic(err)
		}
	}()
	return cmd.Run()
}

func (s *internal) TestSendSimple() {
	s.EqualError(send("should-not-exist", 0, "", true, true), enoent)
	s.EqualError(send(s.pool+"/a/2@snap1", 42, "", true, true), ebadf)
	s.EqualError(send(s.pool+"/"+longName, 42, "", true, true), einval) // WANTE(ENAMETOOLONG)

	// expect epipe
	reader, writer, err := os.Pipe()
	s.Require().NoError(err)
	s.NoError(reader.Close())
	s.EqualError(send(s.pool+"/a/2@snap1", writer.Fd(), "", true, true), epipe)
	s.NoError(writer.Close())

	// expect ebadf
	reader, writer, err = os.Pipe()
	s.Require().NoError(err)
	s.NoError(writer.Close())
	s.EqualError(send(s.pool+"/a/2@snap1", writer.Fd(), "", true, true), ebadf)
	s.NoError(reader.Close())

	// ok
	reader, writer, err = os.Pipe()
	s.Require().NoError(err)
	s.NoError(send(s.pool+"/a/4", writer.Fd(), "", true, true))
	s.NoError(s.receive(reader, s.pool+"/c"))
	s.NoError(reader.Close())
	s.NoError(writer.Close())
}

// Tests both incremental and verifies actual file changes
func (s *internal) TestSendComplex() {
	name := "/" + s.pool + "/a/2/markerfile"
	NoError := s.Require().NoError

	NoError(touch(name))
	NoError(snap(s.pool + "/a/2@pre"))
	NoError(command("sudo", "rm", name).Run())
	NoError(snap(s.pool + "/a/2@post"))

	reader, writer, err := os.Pipe()
	NoError(err)
	NoError(send(s.pool+"/a/2@pre", writer.Fd(), "", true, true))
	s.NoError(writer.Close()) // must be before receiver, otherwise can block
	s.NoError(s.receive(reader, s.pool+"/c"))
	s.NoError(reader.Close())
	name = strings.Replace(name, "/a/", "/c/", 1)
	_, err = os.Stat(name)
	s.NoError(err)

	reader, writer, err = os.Pipe()
	NoError(err)
	s.NoError(send(s.pool+"/a/2@post", writer.Fd(), s.pool+"/a/2@pre", false, false))
	s.NoError(writer.Close()) // must be before receiver, otherwise can block
	s.NoError(s.receive(reader, s.pool+"/c"))
	s.NoError(reader.Close())
	_, err = os.Stat(name)
	s.IsType(&os.PathError{}, err)
}

func (s *internal) TestSnapshot() {
	type errs map[string]syscall.Errno
	type props map[string]string
	tests := []struct {
		desc  string
		pool  string
		snaps []string
		props props
		err   string
		errs  errs
	}{
		{
			desc: "non existent pool",
			pool: "non-existent-pool",
			err:  enoent,
		},
		{
			desc:  "non existent fs",
			pool:  s.pool,
			snaps: []string{s.pool + "/non-existent-fs@1"},
			err:   enoent,
		},
		{
			desc:  "existing snapshot",
			pool:  s.pool,
			snaps: []string{s.pool + "/a/1@snap1"},
			err:   eexist,
			errs:  errs{s.pool + "/a/1@snap1": syscall.EEXIST},
		},
		{
			desc:  "too long snap name",
			pool:  s.pool,
			snaps: []string{s.pool + "/a/1@snap" + longName},
			err:   einval, // WANTE ENAMETOOLONG
		},
		{
			desc:  "too long property name",
			pool:  s.pool,
			snaps: []string{s.pool + "/a/1@snapZ"},
			props: props{longName: "true"},
			err:   einval, // WANTE ENAMETOOLONG
		},
		{
			desc:  "too long property value",
			pool:  s.pool,
			snaps: []string{s.pool + "/a/1@snapZ"},
			props: props{"aprop": strings.Repeat("a", 8192)},
			err:   einval, // WANTE E2BIG
		},
		{
			desc: "multiple snapshots on one fs",
			pool: s.pool,
			snaps: []string{
				s.pool + "/a/1@snapY",
				s.pool + "/a/1@snapZ",
			},
			err: exdev,
		},
		{
			desc: "one valid snapshot",
			pool: s.pool,
			snaps: []string{
				s.pool + "@snapX",
			},
		},
		{
			desc: "multiple valid snapshots",
			pool: s.pool,
			snaps: []string{
				s.pool + "/a@snapY",
				s.pool + "/b@snapY",
				s.pool + "/c@snapY",
			},
		},
	}
	for _, test := range tests {
		s.T().Log(test.desc)
		errs, err := snapshot(test.pool, test.snaps, test.props)
		if test.err == "" {
			s.NoError(err)
		} else {
			s.EqualError(err, test.err)
		}
		if test.errs == nil {
			s.Nil(errs)
		} else {
			s.Equal(map[string]syscall.Errno(test.errs), errs)
		}
	}
}
