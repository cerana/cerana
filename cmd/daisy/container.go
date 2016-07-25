package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/seccomp"
)

type Namespace struct {
	Path string
	Type string
}

type Namespaces []Namespace

// imported from github.com:opencontainers/runc/libcontainer/configs/namespaces_syscall.go
var namespaceInfo = map[string]int{
	"net":   syscall.CLONE_NEWNET,
	"mount": syscall.CLONE_NEWNS,
	"user":  syscall.CLONE_NEWUSER,
	"ipc":   syscall.CLONE_NEWIPC,
	"uts":   syscall.CLONE_NEWUTS,
	"pid":   syscall.CLONE_NEWPID,
	//NEWDATASET:  syscall.CLONE_NEWDATASET,
}

// CloneFlags parses the container's Namespaces options to set the correct
// flags on clone, unshare. This function returns flags only for new namespaces.
func (n Namespaces) CloneFlags() uintptr {
	var flag int
	for _, v := range n {
		if v.Path != "" {
			continue
		}
		flag |= namespaceInfo[v.Type]
	}
	return uintptr(flag)
}

type Container struct {
	Args       []string
	Namespaces Namespaces
	Seccomp    []seccomp.SyscallRule
	Uid        int
	Gid        int
}

func (c *Container) Start() error {
	cmd := &exec.Cmd{
		Path: os.Args[0],
		Args: append([]string{"child"}, c.Args...),
	}
	flags := c.Namespaces.CloneFlags()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: flags,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      c.Uid,
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      c.Gid,
				Size:        1,
			},
		},
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	log.Debugf("child PID: %d", cmd.Process.Pid)
	return cmd.Wait()
}

type Mount struct {
	Source string
	Target string
	Fs     string
	Flags  int
	Data   string
}

type Cfg struct {
	Path     string
	Args     []string
	Hostname string
	Mounts   []Mount
	Rootfs   string
	Devices  []string
}

var defaultMountFlags = syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

var defaultCfg = Cfg{
	Hostname: "daisy",
	Mounts: []Mount{
		{
			Source: "proc",
			Target: "/proc",
			Fs:     "proc",
			Flags:  defaultMountFlags,
		},
		{
			Source: "tmpfs",
			Target: "/dev",
			Fs:     "tmpfs",
			Flags:  syscall.MS_NOSUID | syscall.MS_STRICTATIME,
			Data:   "mode=755",
		},
	},
	Rootfs:  "",
	Devices: []string{"/dev/null", "/dev/zero", "/dev/full", "/dev/random", "/dev/urandom", "/dev/tty", "/dev/ptmx", "/dev/zfs"},
}

func pivotRoot(rootfs string, pivotBaseDir string) (err error) {
	if pivotBaseDir == "" {
		pivotBaseDir = "/"
	}
	// "new_root and put_old must not be on the same filesystem as the current root"
	if err := syscall.Mount(rootfs, rootfs, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
	    return fmt.Errorf("Mount rootfs to itself error: %v", err)
	}
	tmpDir := filepath.Join(rootfs, pivotBaseDir)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("can't create tmp dir %s, error %v", tmpDir, err)
	}
	pivotDir := filepath.Join(tmpDir, ".pivot_root")
	if err := os.MkdirAll(pivotDir, 0755); err != nil {
		return fmt.Errorf("can't create pivot_root dir %s, error %v", pivotDir, err)
	}
	defer func() {
		errVal := os.Remove(pivotDir)
		if err == nil {
			err = errVal
		}
	}()
	if err := syscall.PivotRoot(rootfs, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %s %s", pivotDir, err)
	}
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %s", err)
	}
	// path to pivot dir now changed, update
	pivotDir = filepath.Join(pivotBaseDir, filepath.Base(pivotDir))

	// Make pivotDir rprivate to make sure any of the unmounts don't
	// propagate to parent.
	if err := syscall.Mount("", pivotDir, "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return err
	}

	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %s", err)
	}
	return nil
}

func mount(cfg Cfg) error {
	for _, m := range cfg.Mounts {
		target := filepath.Join(cfg.Rootfs, m.Target)
		log.Debugf("Mount %s to %s", m.Source, target)
		if err := syscall.Mount(m.Source, target, m.Fs, uintptr(m.Flags), m.Data); err != nil {
			return fmt.Errorf("failed to mount %s to %s: %v", m.Source, target, err)
		}
	}
	return nil
}
func createDevices(cfg Cfg) error {
	oldMask := syscall.Umask(0000)
	for _, node := range cfg.Devices {
		var null syscall.Stat_t
		if err := syscall.Stat(node, &null); err != nil {
			continue
		}
		// containers running in a user namespace are not allowed to mknod
		// devices so we can just bind mount it from the host.
		dest := filepath.Join(cfg.Rootfs, node)
		log.Debugf("Mount %s to %s", node, dest)
		if err := bindMountDeviceNode(dest, node); err != nil {
			syscall.Umask(oldMask)
			return err
		}
	}
	syscall.Umask(oldMask)
	return nil
}

func bindMountDeviceNode(dest string, path string) error {
	f, err := os.Create(dest)
	if err != nil && !os.IsExist(err) {
		return err
	}
	if f != nil {
		f.Close()
	}
	return syscall.Mount(path, dest, "bind", syscall.MS_BIND, "")
}

func fixStdioPermissions(uid int, gid int) error {
	var null syscall.Stat_t
	if err := syscall.Stat("/dev/null", &null); err != nil {
		return err
	}
	for _, fd := range []uintptr{
		os.Stdin.Fd(),
		os.Stderr.Fd(),
		os.Stdout.Fd(),
	} {
		var s syscall.Stat_t
		if err := syscall.Fstat(int(fd), &s); err != nil {
			return err
		}
		// skip chown of /dev/null if it was used as one of the STDIO fds.
		if s.Rdev == null.Rdev {
			continue
		}
		if err := syscall.Fchown(int(fd), uid, gid); err != nil {
			return err
		}
	}
	return nil
}

func setup(cfg Cfg) error {
	if err := syscall.Sethostname([]byte(cfg.Hostname)); err != nil {
		return fmt.Errorf("Sethostname: %v", err)
	}
	if err := mount(cfg); err != nil {
		return err
	}
	if err := syscall.Chdir(cfg.Rootfs); err != nil {
		return err
	}
	if err := createDevices(cfg); err != nil {
		return err
	}
	if err := pivotRoot(cfg.Rootfs, ""); err != nil {
		return fmt.Errorf("Pivot root error: %v", err)
	}
	return nil
}

func execProc(cfg Cfg) error {
	log.Debugf("Execute %s", append([]string{cfg.Path}, cfg.Args[1:]...))
	return syscall.Exec(cfg.Path, cfg.Args, os.Environ())
}

func fillCfg() error {
	name, err := exec.LookPath(os.Args[1])
	if err != nil {
		return fmt.Errorf("LookPath: %v", err)
	}
	defaultCfg.Path = name
	defaultCfg.Args = os.Args[1:]
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Error get working dir: %v", err)
	}
	defaultCfg.Rootfs = wd
	return nil
}

func child() error {
	log.Debug("Start child")
	if err := fillCfg(); err != nil {
		return fmt.Errorf("fillCfg: %v", err)
	}
	if err := setup(defaultCfg); err != nil {
		return fmt.Errorf("setup: %v", err)
	}
	if err := SetNoNewPrivs(1); err != nil {
		return fmt.Errorf("SetNoNewPrivs: %v", err)
	}
	if err := SetSubreaper(1); err != nil {
		return fmt.Errorf("SetSubreaper: %v", err)
	}
	if err := seccomp.InitSeccomp(seccomp.Whitelist, DefScmp); err != nil {
		return err
	}

	capInit()

	w, err := newCapWhitelist(Capabilities)
	if err != nil {
		return err
	}
	if err := w.dropBoundingSet(); err != nil {
		return fmt.Errorf("dropBoundingSet: %v", err)
	}

	return execProc(defaultCfg)
}
