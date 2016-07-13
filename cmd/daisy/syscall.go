// +build linux

// imported from github.com:opencontainers/runc/libcontainer/system/linux.go

package main

/*
#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <fcntl.h>
#include <sched.h>

struct nsmap {
	const char *path;
	int type;
};

struct nsmap mappings[]  = {
	{ "MNTNS_PATH", CLONE_NEWNS },
	{ "UTSNS_PATH", CLONE_NEWUTS },
	{ "IPCNS_PATH", CLONE_NEWIPC },
	{ "NETNS_PATH", CLONE_NEWNET },
	{ "PIDNS_PATH", CLONE_NEWPID },
	{ "USERNS_PATH", CLONE_NEWUSER },
	{ NULL, 0 }
};

__attribute__((constructor)) void
init_namespaces(void)
{
	int flags = CLONE_NEWNS|CLONE_NEWUTS|CLONE_NEWIPC|CLONE_NEWNET| \
	    CLONE_NEWPID;
	int fd, err, i;
	const char *nspath;

	(void) unshare(CLONE_NEWUSER);

	for (i = 0; mappings[i].path != NULL; i++) {
		nspath = getenv(mappings[i].path);
		if (nspath != NULL) {
			fd = open(nspath, O_RDONLY, 0644);
			if (fd < 0)
				continue;
   			err = setns(fd, mappings[i].type);
			if (err == 0)
				flags &= ~(mappings[i].type);
			(void) close(fd);
		}
	}

	(void) unshare(flags);
}
*/
import "C"

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/syndtr/gocapability/capability"
)

// If arg2 is nonzero, set the "child subreaper" attribute of the
// calling process; if arg2 is zero, unset the attribute.  When a
// process is marked as a child subreaper, all of the children
// that it creates, and their descendants, will be marked as
// having a subreaper.  In effect, a subreaper fulfills the role
// of init(1) for its descendant processes.  Upon termination of
// a process that is orphaned (i.e., its immediate parent has
// already terminated) and marked as having a subreaper, the
// nearest still living ancestor subreaper will receive a SIGCHLD
// signal and be able to wait(2) on the process to discover its
// termination status.
const PR_SET_CHILD_SUBREAPER = 36
const PR_SET_NO_NEW_PRIVS = 38

type ParentDeathSignal int

func (p ParentDeathSignal) Restore() error {
	if p == 0 {
		return nil
	}
	current, err := GetParentDeathSignal()
	if err != nil {
		return err
	}
	if p == current {
		return nil
	}
	return p.Set()
}

func (p ParentDeathSignal) Set() error {
	return SetParentDeathSignal(uintptr(p))
}

func Execv(cmd string, args []string, env []string) error {
	name, err := exec.LookPath(cmd)
	if err != nil {
		return err
	}

	return syscall.Exec(name, args, env)
}

func Prlimit(pid, resource int, limit syscall.Rlimit) error {
	_, _, err := syscall.RawSyscall6(syscall.SYS_PRLIMIT64, uintptr(pid), uintptr(resource), uintptr(unsafe.Pointer(&limit)), uintptr(unsafe.Pointer(&limit)), 0, 0)
	if err != 0 {
		return err
	}
	return nil
}

func SetParentDeathSignal(sig uintptr) error {
	if _, _, err := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_SET_PDEATHSIG, sig, 0); err != 0 {
		return err
	}
	return nil
}

func GetParentDeathSignal() (ParentDeathSignal, error) {
	var sig int
	_, _, err := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_GET_PDEATHSIG, uintptr(unsafe.Pointer(&sig)), 0)
	if err != 0 {
		return -1, err
	}
	return ParentDeathSignal(sig), nil
}

func SetKeepCaps() error {
	if _, _, err := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_SET_KEEPCAPS, 1, 0); err != 0 {
		return err
	}

	return nil
}

func ClearKeepCaps() error {
	if _, _, err := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_SET_KEEPCAPS, 0, 0); err != 0 {
		return err
	}

	return nil
}

func Setctty() error {
	if _, _, err := syscall.RawSyscall(syscall.SYS_IOCTL, 0, uintptr(syscall.TIOCSCTTY), 0); err != 0 {
		return err
	}
	return nil
}

/*
 * Detect whether we are currently running in a user namespace.
 * Copied from github.com/lxc/lxd/shared/util.go
 */
func RunningInUserNS() bool {
	file, err := os.Open("/proc/self/uid_map")
	if err != nil {
		/*
		 * This kernel-provided file only exists if user namespaces are
		 * supported
		 */
		return false
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	l, _, err := buf.ReadLine()
	if err != nil {
		return false
	}

	line := string(l)
	var a, b, c int64
	fmt.Sscanf(line, "%d %d %d", &a, &b, &c)
	/*
	 * We assume we are in the initial user namespace if we have a full
	 * range - 4294967295 uids starting at uid 0.
	 */
	if a == 0 && b == 0 && c == 4294967295 {
		return false
	}
	return true
}

// SetSubreaper sets the value i as the subreaper setting for the calling process
func SetSubreaper(i int) error {
	return Prctl(PR_SET_CHILD_SUBREAPER, uintptr(i), 0, 0, 0)
}

// Set the no_new_privs bit
func SetNoNewPrivs(i int) error {
	return Prctl(PR_SET_NO_NEW_PRIVS, uintptr(i), 0, 0, 0)
}

func pivotRoot(rootfs, pivotBaseDir string) (err error) {
	if pivotBaseDir == "" {
		pivotBaseDir = "/"
	}
	tmpDir := filepath.Join(rootfs, pivotBaseDir)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("can't create tmp dir %s, error %v", tmpDir, err)
	}
	pivotDir, err := ioutil.TempDir(tmpDir, ".pivot_root")
	if err != nil {
		return fmt.Errorf("can't create pivot_root dir %s, error %v", pivotDir, err)
	}
	defer func() {
		errVal := os.Remove(pivotDir)
		if err == nil {
			err = errVal
		}
	}()
	if err := syscall.PivotRoot(rootfs, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %s", err)
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

func SetNewRoot(path string) (err error) {
	if len(path) == 0 || path == "/" {
		path, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	if err := syscall.Chdir(path); err != nil {
		return err
	}
	if err := createDevices(path); err != nil {
		return err
	}
	return pivotRoot(".", "")
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

// Setuid sets the uid of the calling thread to the specified uid.
func Setuid(uid int) (err error) {
	_, _, e1 := syscall.RawSyscall(syscall.SYS_SETUID, uintptr(uid), 0, 0)
	if e1 != 0 {
		err = e1
	}
	return
}

// Setgid sets the gid of the calling thread to the specified gid.
func Setgid(gid int) (err error) {
	_, _, e1 := syscall.RawSyscall(syscall.SYS_SETGID, uintptr(gid), 0, 0)
	if e1 != 0 {
		err = e1
	}
	return
}

func SetNewUser(uid int, gid int) error {
	if err := fixStdioPermissions(uid, gid); err != nil {
		return err
	}
	if err := Setgid(gid); err != nil {
		return err
	}
	if err := Setuid(uid); err != nil {
		return err
	}
	return nil
}

func Prctl(option int, arg2, arg3, arg4, arg5 uintptr) (err error) {
	_, _, e1 := syscall.Syscall6(syscall.SYS_PRCTL, uintptr(option), arg2, arg3, arg4, arg5, 0)
	if e1 != 0 {
		err = e1
	}
	return
}

var Devices []string

func createDevices(rootfs string) error {
	oldMask := syscall.Umask(0000)
	for _, node := range Devices {
		// containers running in a user namespace are not allowed to mknod
		// devices so we can just bind mount it from the host.
		dest := filepath.Join(rootfs, node)
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
		if len(v.Path) != 0 {
			continue
		}
		flag |= namespaceInfo[v.Type]
	}
	return uintptr(flag)
}

var Capabilities = []string{
	"CAP_CHOWN",
	"CAP_DAC_OVERRIDE",
	"CAP_FSETID",
	"CAP_FOWNER",
	"CAP_MKNOD",
	"CAP_NET_RAW",
	"CAP_SETGID",
	"CAP_SETUID",
	"CAP_SETFCAP",
	"CAP_SETPCAP",
	"CAP_NET_BIND_SERVICE",
	"CAP_SYS_CHROOT",
	"CAP_KILL",
	"CAP_AUDIT_WRITE",
}

const allCapabilityTypes = capability.CAPS | capability.BOUNDS

var capabilityMap map[string]capability.Cap

func capInit() {
	capabilityMap = make(map[string]capability.Cap)
	last := capability.CAP_LAST_CAP
	// workaround for RHEL6 which has no /proc/sys/kernel/cap_last_cap
	if last == capability.Cap(63) {
		last = capability.CAP_BLOCK_SUSPEND
	}
	for _, cap := range capability.List() {
		if cap > last {
			continue
		}
		capKey := fmt.Sprintf("CAP_%s", strings.ToUpper(cap.String()))
		capabilityMap[capKey] = cap
	}
}

func newCapWhitelist(caps []string) (*whitelist, error) {
	l := []capability.Cap{}
	for _, c := range caps {
		v, ok := capabilityMap[c]
		if !ok {
			return nil, fmt.Errorf("unknown capability %q", c)
		}
		l = append(l, v)
	}
	pid, err := capability.NewPid(os.Getpid())
	if err != nil {
		return nil, err
	}
	return &whitelist{
		keep: l,
		pid:  pid,
	}, nil
}

type whitelist struct {
	pid  capability.Capabilities
	keep []capability.Cap
}

// dropBoundingSet drops the capability bounding set to those specified in the whitelist.
func (w *whitelist) dropBoundingSet() error {
	w.pid.Clear(capability.BOUNDS)
	w.pid.Set(capability.BOUNDS, w.keep...)
	return w.pid.Apply(capability.BOUNDS)
}

// drop drops all capabilities for the current process except those specified in the whitelist.
func (w *whitelist) drop() error {
	w.pid.Clear(allCapabilityTypes)
	w.pid.Set(allCapabilityTypes, w.keep...)
	return w.pid.Apply(allCapabilityTypes)
}
