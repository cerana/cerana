package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	//"cgroup":  syscall.CLONE_NEWCGROUP,
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
	Uid        uint32
	Gid        uint32
}

type Mount struct {
	Source string
	Target string
	Fs     string
	Flags  int
	Data   string
}

type Cfg struct {
	Args         []string
	Env          []string
	Uid          int
	Gid          int
	Hostname     string
	Mounts       []Mount
	Rootfs       string
	Devices      []string
	Capabilities []string
	Seccomp      []seccomp.SyscallRule
}

func (c *Container) Start() error {
	var uidmap []syscall.SysProcIDMap
	var gidmap []syscall.SysProcIDMap
	var environment []string
	var inheritFds []*os.File

	environment = []string{
		fmt.Sprintf("TERM=%s", os.Getenv("TERM")),
		fmt.Sprintf("_CERANA_DAISY_UID=%d", c.Uid),
		fmt.Sprintf("_CERANA_DAISY_GID=%d", c.Gid),
	}

	flags := c.Namespaces.CloneFlags()
	if flags&syscall.CLONE_NEWUSER != 0 {
		uidmap = []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      int(c.Uid),
				Size:        1,
			},
		}
		gidmap = []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      int(c.Gid),
				Size:        1,
			},
		}
	}

	for _, v := range c.Namespaces {
		if v.Path == "" {
			continue
		}
		_, ok := namespaceInfo[v.Type]
		if !ok {
			continue
		}
		f, err := os.Open(v.Path)
		if err != nil {
			return fmt.Errorf("Open %s namespace at '%s' failed: %v", v.Type, v.Path, err)
		}
		environment = append(environment, fmt.Sprintf("_CERANA_DAISY_NAMESPACE_%s=%d", v.Type, f.Fd()))
		inheritFds = append(inheritFds, f)
	}

	cmd := &exec.Cmd{
		Path: "/proc/self/exe",
		Args: append([]string{"child"}, os.Args[1:]...),
		Env:  environment,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = inheritFds
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:  flags,
		UidMappings: uidmap,
		GidMappings: gidmap,
		Credential: &syscall.Credential{
			Uid: c.Uid,
			Gid: c.Gid,
		},
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	log.Debugf("child PID: %d", cmd.Process.Pid)
	return cmd.Wait()
}

var defaultMountFlags = syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

func setupRootDir(rootfs string) (err error) {
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount old root as private error: %v", err)
	}
	// "new_root and put_old must not be on the same filesystem as the current root"
	if err := syscall.Mount(rootfs, rootfs, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount rootfs to itself error: %v", err)
	}
	return nil
}

func pivotRoot(rootfs string, pivotBaseDir string) (err error) {
	if pivotBaseDir == "" {
		pivotBaseDir = "/"
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
	//if err := syscall.PivotRoot(rootfs, pivotDir); err != nil {
	if err := syscall.Chroot(rootfs); err != nil {
		return fmt.Errorf("pivot_root %s %s", pivotDir, err)
	}
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %s", err)
	}
	// path to pivot dir now changed, update
	pivotDir = filepath.Join(pivotBaseDir, filepath.Base(pivotDir))

	// Make pivotDir rprivate to make sure any of the unmounts don't
	// propagate to parent.
	//if err := syscall.Mount("", pivotDir, "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
	//	return err
	//}

	//if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
	//	return fmt.Errorf("unmount pivot_root dir %s", err)
	//}
	return nil
}

func mountFilesystems(cfg Cfg) error {
	for _, m := range cfg.Mounts {
		target := filepath.Join(cfg.Rootfs, m.Target)
		log.Debugf("Mount %s to %s", m.Source, target)
		if err := os.MkdirAll(target, 0755); err != nil {
			return fmt.Errorf("cannot create mountpoint %s, error %v", target, err)
		}
		if err := syscall.Mount(m.Source, target, m.Fs, uintptr(m.Flags), m.Data); err != nil {
			return fmt.Errorf("failed to mount %s to %s: %v", m.Source, target, err)
		}
		if m.Flags&syscall.MS_BIND != 0 && m.Flags&syscall.MS_RDONLY != 0 {
			// Some flags only valid for remount
			syscall.Mount(target, target, m.Fs, uintptr(m.Flags|syscall.MS_REMOUNT), m.Data)
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
		if node == "/dev/ptmx" {
			// ptmx must be from same fs as our devpts
			node = filepath.Join(cfg.Rootfs, "/dev/pts/ptmx")
		}
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
	//dieOnError(makeRequest(coordinator, 'getExtNamespaces', httpAddr))

	//select {
	//case err := <-respErr:
	//	dieOnError(err)
	//case result := <-result:
	//	j, _ := json.Marshal(result)
	//	fmt.Println(string(j))
	//}
	for k, v := range namespaceInfo {
		env := os.Getenv(fmt.Sprintf("_CERANA_DAISY_NAMESPACE_%s", k))
		if env == "" {
			continue
		}
		fd, err := strconv.Atoi(env)
		if err != nil {
			log.Fatalf("Invalid child environment")
		}
		err = Setns(uintptr(fd), uintptr(v))
		if err != nil {
			log.Fatalf("Setns: %v", err)
		}
	}
	if err := syscall.Sethostname([]byte(cfg.Hostname)); err != nil {
		return fmt.Errorf("Sethostname: %v", err)
	}
	if err := syscall.Chdir(cfg.Rootfs); err != nil {
		return fmt.Errorf("Cannot enter directory '%s': %v", cfg.Rootfs, err)
	}
	if err := setupRootDir(cfg.Rootfs); err != nil {
		return fmt.Errorf("Cannot set up root filesystem: %v", err)
	}
	if err := mountFilesystems(cfg); err != nil {
		return fmt.Errorf("Cannot mount child filesystems: %v", err)
	}
	if err := createDevices(cfg); err != nil {
		return fmt.Errorf("Cannot create device nodes: %v", err)
	}
	if err := pivotRoot(cfg.Rootfs, ""); err != nil {
		return fmt.Errorf("Pivot root error: %v", err)
	}
	return nil
}

func execProc(cfg Cfg) error {
	path := filepath.Join("/", cfg.Args[0])
	log.Debugf("Execute %s", append([]string{path}, cfg.Args...))
	return syscall.Exec(path, cfg.Args, cfg.Env)
}

func fillCfg(cfg Cfg) error {
	uid, err := strconv.Atoi(os.Getenv("_CERANA_DAISY_UID"))
	if err != nil {
		log.Fatalf("Invalid child environment")
	}
	gid, err := strconv.Atoi(os.Getenv("_CERANA_DAISY_GID"))
	if err != nil {
		log.Fatalf("Invalid child environment")
	}
	cfg.Uid = uid
	cfg.Gid = gid

	return nil
}

func Child(cfg Cfg) error {
	log.Debug("Start child")
	if err := fillCfg(cfg); err != nil {
		return fmt.Errorf("fillCfg: %v", err)
	}
	if err := setup(cfg); err != nil {
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

	w, err := newCapWhitelist(cfg.Capabilities)
	if err != nil {
		return fmt.Errorf("newCapWhitelist: %v", err)
	}

	// drop capabilities in bounding set before changing user
	if err := w.dropBoundingSet(); err != nil {
		return fmt.Errorf("dropBoundingSet: %v", err)
	}

	// preserve existing capabilities while we change users
	if err := SetKeepCaps(); err != nil {
		return fmt.Errorf("SetKeepCaps: %v", err)
	}

	if err := SetNewUser(0, 0); err != nil {
		return fmt.Errorf("SetNewUser: %v", err)
	}

	if err := ClearKeepCaps(); err != nil {
		return fmt.Errorf("ClearKeepCaps: %v", err)
	}

	//dieOnError(selinux.InitLabels(nil));
	return execProc(cfg)
}
