// +build linux

// This is the daisy binary to execute the target program with isolation applied
package main

import (
	"fmt"
	"math/rand"
	"os"
	//"runtime"
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

func init() {
	// pin main goroutine to thread
	//runtime.LockOSThread()
	log.SetLevel(log.DebugLevel)
	rand.Seed(time.Now().UnixNano())
}

func main() {

	var coordinator, namespaces, extNamespaces, rootFs, hostname, devices, env string
	var uid, gid, uidrange, gidrange int
	var kvm bool
	var execArgs []string
	var nsList Namespaces
	var envList []string

	flags.StringVarP(&rootFs, "root directory", "r", "", "location of the container root")
	flags.IntVarP(&uid, "uid", "u", os.Getuid(), "user id to use as base")
	flags.IntVarP(&gid, "gid", "g", os.Getgid(), "group id to use as base")
	flags.IntVarP(&uidrange, "uid range", "U", 1000, "length of mapped user id range")
	flags.IntVarP(&gidrange, "gid range", "G", 1000, "length of mapped group id range")
	flags.BoolVarP(&kvm, "kvm mode", "k", false, "whether we are running just qemu")
	flags.StringVarP(&env, "environment", "e", "", "list of environment variables to set")
	flags.StringVarP(&hostname, "hostname", "h", "daisy", "hostname of new uts namespace")
	flags.StringVarP(&devices, "devices", "d", "null,zero,full,random,urandom,tty,ptmx,zfs", "list of device nodes to allow")
	flags.StringVarP(&namespaces, "namespace list", "n", "user,mount,uts,pid,ipc", "list of namespaces to unshare")
	flags.StringVarP(&extNamespaces, "external namespace", "x", "", fmt.Sprintf("list of external namespaces of the form'type%spath'", ArgSep))
	flags.StringVarP(&coordinator, "coordinator_url", "c", "", "url of the coordinator")
	flags.Parse()

	// extract argv for executable
	execArgs = flags.Args()
	if len(execArgs) == 0 {
		log.Fatalf("Missing path to executable")
		os.Exit(1)
	}
	envList = strings.Split(env, ",")
	for _, ns := range strings.Split(namespaces, ",") {
		nsList = append(nsList, Namespace{Type: ns, Path: ""})
	}

	if os.Args[0] == "child" {
		var devList []string
		var scmp = seccomp.Whitelist

		if kvm {
			scmp = seccomp.WhitelistKVM
		}

		cfg := defaultCfg
		cfg.Args = execArgs
		cfg.Env = envList
		cfg.Rootfs = rootFs
		cfg.Hostname = hostname
		cfg.Seccomp = scmp

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

// argmap is an arbitrarilly nested map
// copied from coordinater-cli
type argmap map[string]interface{}

func (am argmap) set(keys []string, value interface{}) error {
	if len(keys) == 0 {
		return nil
	}

	key := keys[0]

	if len(keys) == 1 {
		am[key] = value
		return nil
	}

	var m argmap
	mi, ok := am[key]
	if !ok {
		m = make(argmap)
		am[key] = m
	} else {
		m, ok = mi.(argmap)
		if !ok {
			return fmt.Errorf("intermediate nested key %s defined and not a map", key)
		}
	}

	return m.set(keys[1:], value)
}

func parseArgs(args []string, argSep string) (map[string]interface{}, error) {
	out := make(argmap)
	for _, in := range args {
		parts := strings.Split(in, argSep)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid request arg: '%s'", in)
		}

		valueS := strings.Join(parts[1:], argSep)
		var value interface{}
		if arg, err := strconv.ParseInt(valueS, 10, 64); err == nil {
			value = arg
		} else if arg, err := strconv.ParseBool(valueS); err == nil {
			value = arg
		} else {
			value = valueS
		}

		keys := strings.Split(parts[0], ".")
		if err := out.set(keys, value); err != nil {
			return nil, err
		}

	}
	return out, nil
}
