package kv

import (
	"os"
	"os/exec"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/internal/tests/common"
	"github.com/cerana/cerana/provider"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
func NewMock(config *provider.Config, tracker *acomm.Tracker) (*Mock, error) {
	s := common.Suite{}
	s.KVPrefix = "mock-kv-provider"
	s.KVCmdMaker = common.ConsulMaker
	s.SetupSuite()

	v := viper.New()
	flagset := pflag.NewFlagSet("mock-kv-provider", pflag.PanicOnError)
	conf := NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse(nil))
	v.Set("service_name", "mock-kv-provider")
	v.Set("socket_dir", config.SocketDir())
	v.Set("coordinator_url", config.CoordinatorURL())
	v.Set("log_level", "fatal")
	v.Set("address", s.KVURL)
	s.Require().NoError(config.LoadConfig())
	s.Require().NoError(config.SetupLogging())

	kv, err := New(conf, tracker)
	if err != nil {
		return nil, err
	}

	return &Mock{KV: kv, cmd: s.KVCmd, dir: s.KVDir}, nil
}

// Stop will stop the kv and remove the temporary directory used for it's data
func (m *Mock) Stop() {
	_ = m.cmd.Process.Kill()
	_ = m.cmd.Wait()
	_ = os.RemoveAll(m.dir)
}

// Get will perform a Get operation directly on the kv store.
func (m *Mock) Get(key string) (Value, error) {
	if m.kvDown() {
		return Value{}, errorKVDown
	}

	kvV, err := m.KV.kv.Get(key)
	return Value(kvV), err
}

// Set will perform a Set operation directly on the kv store.
func (m *Mock) Set(key, value string) error {
	if m.kvDown() {
		return errorKVDown
	}

	return m.KV.kv.Set(key, value)
}

// Clean will perform a recursive Delete operation directly on the kv store.
func (m *Mock) Clean(prefix string) error {
	if m.kvDown() {
		return errorKVDown
	}

	return m.KV.kv.Delete(prefix, true)
}
