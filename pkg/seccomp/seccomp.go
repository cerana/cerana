// +build linux,cgo

// Package seccomp provides basic interfaces for applying seccomp filters
package seccomp

// from github.com/opencontainers/runc/libcontainer/seccomp/seccomp_linux.go

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	libseccomp "github.com/seccomp/libseccomp-golang"
)

// ArgFilter describes a filter expression to match syscall arguments
type ArgFilter struct {
	arg    uint
	op     string
	value1 uint64
	value2 uint64
}

// SyscallRule is a single seccomp action for a particular syscall and args
type SyscallRule struct {
	name   string
	args   []ArgFilter
	action string
}

var (
	actAllow = libseccomp.ActAllow
	actTrap  = libseccomp.ActTrap
	actKill  = libseccomp.ActKill
	actTrace = libseccomp.ActTrace.SetReturnCode(int16(syscall.EPERM))
	actErrno = libseccomp.ActErrno.SetReturnCode(int16(syscall.EPERM))

	// SeccompModeFilter refers to the syscall argument SECCOMP_MODE_FILTER.
	SeccompModeFilter = uintptr(2)
	actions           = map[string]libseccomp.ScmpAction{
		"SCMP_ACT_TRAP":  actTrap,
		"SCMP_ACT_KILL":  actKill,
		"SCMP_ACT_ALLOW": actAllow,
		"SCMP_ACT_TRACE": actTrace,
		"SCMP_ACT_ERRNO": actErrno,
	}
	operators = map[string]libseccomp.ScmpCompareOp{
		"==": libseccomp.CompareEqual,
		"!=": libseccomp.CompareNotEqual,
		">":  libseccomp.CompareGreater,
		">=": libseccomp.CompareGreaterEqual,
		"<":  libseccomp.CompareLess,
		"<=": libseccomp.CompareLessOrEqual,
		"~=": libseccomp.CompareMaskedEqual,
	}
)

// Filters given syscalls in a container, preventing them from being used
// Started in the container init process, and carried over to all child processes

// InitSeccomp loads the specified policy of rules and a default action
func InitSeccomp(rules []SyscallRule, defaultAction string) error {
	filter, err := libseccomp.NewFilter(actions[defaultAction])

	if len(rules) == 0 {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error creating filter: %s", err)
	}

	// Unset no new privs bit
	if err = filter.SetNoNewPrivsBit(false); err != nil {
		return fmt.Errorf("error setting no new privileges: %s", err)
	}

	// Add a rule for each syscall
	for _, call := range rules {
		if err = matchCall(filter, call.name, call.args, call.action); err != nil {
			return err
		}
	}

	if err = filter.Load(); err != nil {
		return fmt.Errorf("error loading seccomp filter into kernel: %s", err)
	}

	return nil
}

// IsEnabled returns true if the kernel has been configured to support seccomp
func IsEnabled() bool {
	// Try to read from /proc/self/status for kernels > 3.8
	s, err := parseStatusFile("/proc/self/status")
	if err != nil {
		// Check if Seccomp is supported, via CONFIG_SECCOMP.
		if _, _, err := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_GET_SECCOMP, 0, 0); err != syscall.EINVAL {
			// Make sure the kernel has CONFIG_SECCOMP_FILTER.
			if _, _, err := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_SET_SECCOMP, SeccompModeFilter, 0); err != syscall.EINVAL {
				return true
			}
		}
		return false
	}
	_, ok := s["Seccomp"]
	return ok
}

// Convert filter to seccomp condition
func getCondition(args ArgFilter) (libseccomp.ScmpCondition, error) {
	cond := libseccomp.ScmpCondition{}

	op, ok := operators[args.op]
	if !ok {
		return cond, fmt.Errorf("invalid operator")
	}

	return libseccomp.MakeCondition(args.arg, op, args.value1, args.value2)
}

// Add a rule to match a single syscall
func matchCall(filter *libseccomp.ScmpFilter, name string, args []ArgFilter, action string) error {
	if filter == nil {
		return fmt.Errorf("nil passed as filter")
	}

	actNum, ok := actions[action]

	if len(name) == 0 || len(action) == 0 || !ok {
		return fmt.Errorf("empty string is not a valid syscall or action")
	}

	// If we can't resolve the syscall, assume it's not supported on this kernel
	// Ignore it, don't error out
	callNum, err := libseccomp.GetSyscallFromName(name)
	if err != nil {
		return nil
	}

	// Unconditional match - just add the rule
	if len(args) == 0 {
		if err = filter.AddRule(callNum, actNum); err != nil {
			return err
		}
	} else {
		// Conditional match - convert the per-arg rules into library format
		conditions := []libseccomp.ScmpCondition{}

		for _, cond := range args {
			newCond, err := getCondition(cond)
			if err != nil {
				return err
			}

			conditions = append(conditions, newCond)
		}

		if err = filter.AddRuleConditional(callNum, actNum, conditions); err != nil {
			return err
		}
	}

	return nil
}

func parseStatusFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	status := make(map[string]string)

	for s.Scan() {
		if err := s.Err(); err != nil {
			return nil, err
		}

		text := s.Text()
		parts := strings.Split(text, ":")

		if len(parts) <= 1 {
			continue
		}

		status[parts[0]] = parts[1]
	}
	return status, nil
}
