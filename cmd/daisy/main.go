// +build linux

// This is the daisy binary to execute the target program with isolation applied
package main

import (
	"fmt"
	"os"
	//"runtime"
	"strconv"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/seccomp"
	flags "github.com/spf13/pflag"
)

const (
	ArgSep  = "="
	DefScmp = "SCMP_ACT_ERRNO"
)

func init() {
	// pin main goroutine to thread
	//runtime.LockOSThread()
	log.SetLevel(log.DebugLevel)
}

func main() {

	var coordinator, namespaces, extNamespaces, rootFs string
	var uid, gid, uidrange, gidrange int
	var kvm bool
	var execArgs []string
	var nsList Namespaces
	var scmp = seccomp.Whitelist

	if os.Args[0] == "child" {
		err := child()
		dieOnError(err)
		os.Exit(1)
	}

	flags.StringVarP(&rootFs, "root directory", "r", "", "location of the container root")
	flags.IntVarP(&uid, "uid", "u", os.Getuid(), "user id to use as base")
	flags.IntVarP(&gid, "gid", "g", os.Getgid(), "group id to use as base")
	flags.IntVarP(&uidrange, "uid range", "U", 1000, "length of mapped user id range")
	flags.IntVarP(&gidrange, "gid range", "G", 1000, "length of mapped group id range")
	flags.BoolVarP(&kvm, "kvm mode", "k", false, "whether we are running just qemu")
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

	for _, ns := range strings.Split(namespaces, ",") {
		nsList = append(nsList, Namespace{Type: ns, Path: ""})
	}

	if kvm {
		scmp = seccomp.WhitelistKVM
	}

	//dieOnError(SetNoNewPrivs(1))

	//dieOnError(seccomp.InitSeccomp(scmp, DefScmp))

	//capInit()

	//w, err := newCapWhitelist(Capabilities)

	//dieOnError(err)

	// drop capabilities in bounding set before changing user
	//dieOnError(w.dropBoundingSet())

	// preserve existing capabilities while we change users
	//dieOnError(SetKeepCaps())

	//dieOnError(SetNewUser(uid, gid))

	//dieOnError(makeRequest(coordinator, 'getExtNamespaces', httpAddr))

	//select {
	//case err := <-respErr:
	//	dieOnError(err)
	//case result := <-result:
	//	j, _ := json.Marshal(result)
	//	fmt.Println(string(j))
	//}

	//dieOnError(selinux.InitLabels(nil));
	// should update user namespace mapping before this point
	if rootFs != "/" && rootFs != "" {
		if err := syscall.Chdir(rootFs); err != nil {
			log.Fatalf("Cannot enter root directory '%s': %v", rootFs, err)
			os.Exit(1)
		}
	}
	c := &Container{
		Args:       execArgs,
		Uid:        uid,
		Gid:        gid,
		Namespaces: nsList,
		Seccomp:    scmp,
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
