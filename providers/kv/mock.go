package kv

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/internal/tests/common"
)

// Mock is a mock KV provider
type Mock struct {
	*KV
	dir string
	cmd *exec.Cmd
}

// NewMock starts up a kv backend server and instantiates a new kv.KV provider.
// The kv backend is started on the port provided as part of config.Address().
// Mock.Stop() should be called when testing is done in order to clean up.
func NewMock(config *Config, tracker *acomm.Tracker) (*Mock, error) {
	dir, err := ioutil.TempDir("", "mock-kv-provider-")
	if err != nil {
		return nil, err
	}

	addr, err := config.Address()
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	_, pStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, err
	}

	port := uint16(0)
	_, err = fmt.Sscanf(pStr, "%d", &port)
	if err != nil {
		return nil, err
	}

	cmd := common.ConsulMaker(port, dir, path.Base(dir))
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	kv, err := New(nil, tracker)
	if err != nil {
		return nil, err
	}

	return &Mock{
		KV:  kv,
		cmd: cmd,
		dir: dir,
	}, nil
}

// Stop will stop the kv and remove the temporary directory used for it's data
func (m *Mock) Stop() {
	_ = m.cmd.Process.Kill()
	_ = m.cmd.Wait()
	_ = os.RemoveAll(m.dir)
}
