package common

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"syscall"
	"testing"

	"gopkg.in/tomb.v2"
)

// Cmd wraps an exec.Cmd with monitoring and easy access to output.
type Cmd struct {
	Cmd *exec.Cmd
	Out *bytes.Buffer
	t   tomb.Tomb
}

// Start runs a command asynchronously.
func Start(cmdName string, args ...string) (*Cmd, error) {
	cmd := exec.Command(cmdName, args...)
	out := &bytes.Buffer{}
	if testing.Verbose() {
		cmd.Stdout = io.MultiWriter(os.Stderr, out)
		cmd.Stderr = io.MultiWriter(os.Stderr, out)
	} else {
		cmd.Stdout = out
		cmd.Stderr = out
	}

	c := &Cmd{
		Cmd: cmd,
		Out: out,
	}

	if err := c.Cmd.Start(); err != nil {
		return c, err
	}

	c.t.Go(c.Cmd.Wait)
	return c, nil
}

// Stop kills a running command and waits until it exits.
// The error returned is from the Kill call, not the error of the exiting command.
// For the latter, call c.Err() after c.Stop().
func (c *Cmd) Stop() error {
	if !c.Alive() {
		return nil
	}
	if err := c.Cmd.Process.Kill(); err != nil {
		return err
	}
	_ = c.t.Wait()
	return nil
}

// Wait waits for a command to finish and returns the exit error.
func (c *Cmd) Wait() error {
	if c.t.Alive() {
		return c.t.Wait()
	}
	return c.t.Err()
}

// Alive returns whether the command process is alive or not.
func (c *Cmd) Alive() bool {
	return c.t.Alive()
}

// ExitStatus returns the exit status code and error for a command.
// If the command is still running or in the process of being shut down, the exit code will be 0 and the returned error will be non-nil.
func (c *Cmd) ExitStatus() (int, error) {
	err := c.t.Err()
	return ExitStatus(err), err
}

// ExecSync runs a command synchronously, waiting for it to complete.
func ExecSync(cmdName string, args ...string) (*Cmd, error) {
	cmd := exec.Command(cmdName, args...)
	c := &Cmd{
		Cmd: cmd,
	}

	var out []byte
	c.t.Go(func() error {
		var err error
		out, err = cmd.CombinedOutput()
		return err
	})

	err := c.Wait()
	c.Out = bytes.NewBuffer(out)
	return c, err
}

// ExitStatus tries to extract an exit status code from an error.
func ExitStatus(err error) int {
	exitStatus := 0
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitStatus = status.ExitStatus()
			}
		}
	}
	return exitStatus
}

// Build builds the current go package.
func Build() error {
	if os.Getenv("LOCHNESS_TEST_NO_BUILD") != "" {
		return nil
	}
	_, err := ExecSync("go", "build", "-i")
	return err
}
