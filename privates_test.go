package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/stretchr/testify/suite"
)

var (
	eexist       = syscall.EEXIST.Error()
	einval       = syscall.EINVAL.Error()
	enametoolong = syscall.ENAMETOOLONG.Error()
	enoent       = syscall.ENOENT.Error()
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
		f, err := ioutil.TempFile("", "gozfs-test-temp")
		if err != nil {
			panic(err)
		}
		files[i] = f.Name()
		f.Close()
	}
	s.files = files

	script := []byte(`
	set -e
	pool=$1
	shift
	zpool list $pool &>/dev/null && zpool destroy $pool
	files=($@)
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

	args := make([]string, 3, 3+len(files))
	args[0] = "bash"
	args[1] = "/dev/stdin"
	args[2] = s.pool
	args = append(args, files...)
	cmd := command("sudo", args...)

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
		os.Remove(s.files[i])
	}
	s.Require().NoError(err)
}

func (s *internal) SetupTest() {
	s.create("gozfs-test")
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
	s.EqualError(create("gozfs-test-no-exists/1", dmuZFS, nil), enoent)
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
	u1024 := uint64(1024)
	u8192 := u1024 * 8

	tests := []struct {
		name  string
		props p
		err   string
	}{
		{"gozfs-test-no-exists/1", nil, enoent},
		{s.pool + "/" + longName, nil, einval}, // WANTE(ENAMETOOLONG)
		{s.pool + "/~1", nil, einval},
		{s.pool + "/1", p{"bad-prop": true}, einval},
		{s.pool + "/2", p{"volsize": 0}, einval},
		{s.pool + "/3", p{"volsize": 1}, einval},
		{s.pool + "/4", p{"volsize": 8*1024 + 1}, einval},
		{s.pool + "/5", p{"volsize": 8 * 1024}, einval},
		{s.pool + "/6", p{"volsize": uint64(0)}, einval},
		{s.pool + "/7", p{"volsize": uint64(1)}, einval},
		{s.pool + "/8", p{"volsize": u8192 + 1}, einval},
		{s.pool + "/9", p{"volsize": u8192}, ""},
		{s.pool + "/9", p{"volsize": u8192}, eexist},
		{s.pool + "/10", p{"volsize": 0, "volblocksize": 1024}, einval},
		{s.pool + "/11", p{"volsize": 1, "volblocksize": 1024}, einval},
		{s.pool + "/12", p{"volsize": 1024 + 1, "volblocksize": 1024}, einval},
		{s.pool + "/13", p{"volsize": 1024, "volblocksize": 1024}, einval},
		{s.pool + "/14", p{"volsize": uint64(0), "volblocksize": 1024}, einval},
		{s.pool + "/15", p{"volsize": uint64(1), "volblocksize": 1024}, einval},
		{s.pool + "/16", p{"volsize": u1024 + 1, "volblocksize": 1024}, einval},
		{s.pool + "/17", p{"volsize": u1024, "volblocksize": 1024}, einval},
		{s.pool + "/18", p{"volsize": 0, "volblocksize": u1024}, einval},
		{s.pool + "/19", p{"volsize": 1, "volblocksize": u1024}, einval},
		{s.pool + "/20", p{"volsize": 1024 + 1, "volblocksize": u1024}, einval},
		{s.pool + "/21", p{"volsize": 1024, "volblocksize": u1024}, einval},
		{s.pool + "/22", p{"volsize": uint64(0), "volblocksize": u1024}, einval},
		{s.pool + "/23", p{"volsize": uint64(1), "volblocksize": u1024}, einval},
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

func (s *internal) TestListEmpty() {
	s.destroy()

	m, err := list("", nil, true, 0)
	s.NoError(err, "m: %v", m)
	s.Assert().Len(m, 0)

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
