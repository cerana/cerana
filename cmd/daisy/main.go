// +build linux

// This is the daisy binary to execute the target program with isolation applied
package main

import "C"

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/seccomp"
	flags "github.com/spf13/pflag"
)

const (
	argSep  = "="
	defScmp = "SCMP_ACT_ERRNO"
)

func main() {
	// pin main goroutine to thread
	runtime.LockOSThread()
	log.SetLevel(log.FatalLevel)

	var coordinator, namespaces, rootDir, binPath string
	var uid, gid, uidrange, gidrange int
	var kvm bool
	var extNamespaces, binArgs []string
	var scmp = seccomp.Whitelist

	flags.StringVarP(&rootDir, "root directory", "r", "/", "location of the container root")
	flags.StringVarP(&binPath, "executable path", "p", "/bin/sh", "relative path and args of the program to exec")
	flags.IntVarP(&uid, "uid", "u", 0, "user id to use as base")
	flags.IntVarP(&gid, "gid", "g", 0, "group id to use as base")
	flags.IntVarP(&uidrange, "uid range", "U", 1000, "length of mapped user id range")
	flags.IntVarP(&gidrange, "gid range", "G", 1000, "length of mapped group id range")
	flags.BoolVarP(&kvm, "kvm mode", "k", false, "whether we are running just qemu")
	flags.StringVarP(&namespaces, "namespace list", "n", "user,mount,uts,pid,ipc", "list of namespaces to unshare")
	flags.StringSliceVarP(&extNamespaces, "external namespace", "x", []string{}, fmt.Sprintf("list of external namespaces of the form'type%spath'. can be set multiple times", argSep))
	flags.StringVarP(&coordinator, "coordinator_url", "c", "", "url of the coordinator")
	flags.Parse()

	// extract argv for executable
	binArgs = strings.Split(binPath, " ")
	binPath = binArgs[0]

	args, err := parseArgs(extNamespaces)
	dieOnError(err)

	if len(args) != 0 {
		fmt.Println("please use *NS_PATH environment variables to specify namespace paths")
	}

	if kvm {
		scmp = seccomp.WhitelistKVM
	}

	dieOnError(SetNoNewPrivs(1))

	dieOnError(SetNewRoot(rootDir))

	dieOnError(SetSubreaper(1))

	dieOnError(seccomp.InitSeccomp(scmp, defScmp))

	capInit()

	w, err := newCapWhitelist(Capabilities)
	dieOnError(err)
	// drop capabilities in bounding set before changing user
	dieOnError(w.dropBoundingSet())

	// preserve existing capabilities while we change users
	dieOnError(SetKeepCaps())

	dieOnError(SetNewUser(uid, gid))

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
	dieOnError(SetNewUser(0, 0))

	dieOnError(ClearKeepCaps())

	// finally, exec user program
	Execv(binPath, binArgs, nil)
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

func parseArgs(args []string) (map[string]interface{}, error) {
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
