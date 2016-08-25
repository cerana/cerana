// +build linux

// This is the daisy binary to execute the target program with isolation applied
package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/seccomp"
	flags "github.com/spf13/pflag"
)

const (
	ArgSep  = "="
	DefScmp = "SCMP_ACT_ERRNO"
)

var defaultCfg = Cfg{
	Hostname: "",
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
		{
			Source: "devpts",
			Target: "/dev/pts",
			Fs:     "devpts",
			Flags:  syscall.MS_NOSUID | syscall.MS_NOEXEC,
			Data:   "newinstance,ptmxmode=0666,mode=660",
		},
		{
			Source: "/sys",
			Target: "/sys",
			Fs:     "bind",
			Flags:  syscall.MS_BIND | syscall.MS_REC | syscall.MS_RDONLY,
		},
	},
	Rootfs:  "",
	Devices: []string{},
}

type stringValue []string

func (s *stringValue) String() string {
	return fmt.Sprint(*s)
}

func (s *stringValue) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (s *stringValue) Type() string {
	return "stringValue"
}

type kvPair struct{
	key string
	value string
}

type kvPairValue []kvPair

func (k *kvPairValue) String() string {
	return fmt.Sprint(*k)
}

func (k *kvPairValue) Set(value string) error {
	for _, kv := range strings.Split(value, ",") {
		parts := strings.Split(kv, ArgSep)
		if len(parts) != 2 {
			return fmt.Errorf("invalid key%svalue pair '%s'", ArgSep, kv)
		}
		*k = append(*k, kvPair{parts[0], parts[1]})
	}
	return nil
}

func (k *kvPairValue) Type() string {
	return "kvPairValue"
}

func init() {
	// pin main goroutine to thread
	runtime.LockOSThread()
	rand.Seed(time.Now().UnixNano())

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
}

func main() {

	var coordinator, namespaces, rootFs, hostname, devices string
	var uid, gid, uidrange, gidrange int
	var verbose, kvm bool
	var execArgs []string
	var nsList Namespaces
	var env stringValue
	var extNamespaces kvPairValue

	flags.BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	flags.StringVarP(&rootFs, "root directory", "r", "", "location of the container root")
	flags.IntVarP(&uid, "uid", "u", os.Getuid(), "user id to use as base")
	flags.IntVarP(&gid, "gid", "g", os.Getgid(), "group id to use as base")
	flags.IntVarP(&uidrange, "uid range", "U", 1000, "length of mapped user id range")
	flags.IntVarP(&gidrange, "gid range", "G", 1000, "length of mapped group id range")
	flags.BoolVarP(&kvm, "kvm mode", "k", false, "whether we are running just qemu")
	flags.VarP(&env, "environment", "e", "environment variable to set in the form 'name=value'")
	flags.StringVarP(&hostname, "hostname", "h", "daisy", "hostname of new uts namespace")
	flags.StringVarP(&devices, "devices", "d", "null,zero,full,random,urandom,tty,ptmx,zfs", "list of device nodes to allow")
	flags.StringVarP(&namespaces, "namespace list", "n", "user,mount,uts,pid,ipc", "list of namespaces to unshare")
	flags.VarP(&extNamespaces, "external namespace", "x", fmt.Sprintf("list of external namespaces in the form 'type%spath'", ArgSep))
	flags.StringVarP(&coordinator, "coordinator_url", "c", "", "url of the coordinator")
	flags.Parse()

	if verbose {
		log.SetLevel(log.DebugLevel)
	}
	// extract argv for executable
	execArgs = flags.Args()
	if len(execArgs) == 0 {
		log.Fatalf("Missing path to executable")
		os.Exit(1)
	}
	for _, ns := range strings.Split(namespaces, ",") {
		nsList = append(nsList, Namespace{Type: ns, Path: ""})
	}
	for _, pair := range extNamespaces {
		nsList = append(nsList, Namespace{Type: pair.key, Path: pair.value})
	}

	if os.Args[0] == "child" {
		var devList []string
		var scmp = seccomp.Whitelist
		var caps = CapabilitiesDefault

		if kvm {
			scmp = seccomp.WhitelistKVM
			caps = CapabilitiesKVM
		}

		cfg := defaultCfg
		cfg.Args = execArgs
		cfg.Env = env
		cfg.Rootfs = rootFs
		cfg.Hostname = hostname
		cfg.Seccomp = scmp
		cfg.Capabilities = caps

		for _, dev := range strings.Split(devices, ",") {
			devList = append(devList, "/dev/"+dev)
		}
		cfg.Devices = devList
		if err := Child(cfg); err != nil {
			log.Fatalf("Child execution failed: %v", err)
		}
		os.Exit(1)
	}

	if rootFs != "/" && rootFs != "" {
		if err := syscall.Chdir(rootFs); err != nil {
			log.Fatalf("Cannot enter root directory '%s': %v", rootFs, err)
			os.Exit(1)
		}
	}

	c := &Container{
		Args:       execArgs,
		Uid:        uint32(uid),
		Gid:        uint32(gid),
		Namespaces: nsList,
	}
	if err := c.Start(); err != nil {
		log.Fatalf("Container start failed: %v", err)
	}
}

func dieOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

