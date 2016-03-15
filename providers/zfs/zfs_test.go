package zfs_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/mistifyio/mistify/acomm"
	"github.com/mistifyio/mistify/provider"
	zfsp "github.com/mistifyio/mistify/providers/zfs"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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

	longName = strings.Repeat("z", 257)
)

type zfs struct {
	suite.Suite
	pool    string
	files   []string
	dir     string
	tracker *acomm.Tracker
	zfs     *zfsp.ZFS
}

func TestZFS(t *testing.T) {
	suite.Run(t, new(zfs))
}

type props map[string]interface{}

func command(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	return cmd
}

func (s *zfs) zfsSetup(pool string) {
	s.pool = pool
	files := make([]string, 5)
	for i := range files {
		f, err := ioutil.TempFile(s.dir, "zpool-img")
		if err != nil {
			panic(err)
		}
		files[i] = f.Name()
		_ = f.Close()
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

	zfs create $pool/fs
	zfs create $pool/fs/1snap
	zfs snapshot $pool/fs/1snap@snap

	zfs create $pool/fs/hold_snap
	zfs snapshot $pool/fs/hold_snap@snap
	zfs hold hold $pool/fs/hold_snap@snap

	zfs create $pool/fs/unmounted
	zfs unmount $pool/fs/unmounted

	zfs create $pool/fs/unmounted_children
	zfs create $pool/fs/unmounted_children/1
	zfs create $pool/fs/unmounted_children/2
	zfs unmount $pool/fs/unmounted_children

	zfs snapshot $pool/fs@snap_with_clone
	zfs clone $pool/fs@snap_with_clone $pool/fs_clone
	zfs unmount $pool/fs_clone


	zfs create $pool/vol
	zfs create -V 8192 $pool/vol/1snap
	zfs snapshot $pool/vol/1snap@snap

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

func (s *zfs) zfsTearDown() {
	err := command("sudo", "zpool", "destroy", s.pool).Run()
	for i := range s.files {
		_ = os.Remove(s.files[i])
	}
	s.Require().NoError(err)
}

func unmount(ds string) error {
	return command("sudo", "zfs", "unmount", ds).Run()
}

func unhold(tag, snapshot string) error {
	return command("sudo", "zfs", "release", tag, snapshot).Run()
}

func (s *zfs) SetupSuite() {
	dir, err := ioutil.TempDir("", "zfs-provider-test-")
	s.Require().NoError(err)
	s.dir = dir

	v := viper.New()
	flagset := pflag.NewFlagSet("go-zfs", pflag.PanicOnError)
	config := provider.NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	v.Set("service_name", "zfs-provider-test")
	v.Set("socket_dir", s.dir)
	v.Set("coordinator_url", "unix:///tmp/foobar")
	v.Set("log_level", "fatal")
	s.Require().NoError(config.LoadConfig())
	s.Require().NoError(config.SetupLogging())

	tracker, err := acomm.NewTracker(filepath.Join(s.dir, "tracker.sock"), nil, 5*time.Second)
	s.Require().NoError(err)
	s.Require().NoError(tracker.Start())
	s.tracker = tracker

	s.zfs = zfsp.New(config, tracker)
}

func (s *zfs) SetupTest() {
	s.zfsSetup("zfs-provider-test")
}

func (s *zfs) TearDownTest() {
	s.zfsTearDown()
}

func (s *zfs) TearDownSuite() {
	s.tracker.Stop()
	_ = os.RemoveAll(s.dir)
}
