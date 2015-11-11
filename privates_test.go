package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
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

func (s *internal) TestDummy() {
}
