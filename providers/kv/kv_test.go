package kv

import (
	"io/ioutil"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/internal/tests/common"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

func TestKV(t *testing.T) {
	suite.Run(t, new(KVS))
}

type KVS struct {
	common.Suite
	config  *Config
	tracker *acomm.Tracker
	KV      *KV
	keys    []string
}

func (s *KVS) SetupSuite() {
	s.Suite.SetupSuite()

	dir, err := ioutil.TempDir("", "kv-provider-test-")
	s.Require().NoError(err)

	v := viper.New()
	flagset := pflag.NewFlagSet("kv-provider", pflag.PanicOnError)
	config := NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	v.Set("service_name", "kv-provider-test")
	v.Set("socket_dir", dir)
	v.Set("coordinator_url", "unix:///tmp/foobar")
	v.Set("log_level", "fatal")
	v.Set("address", s.KVURL)
	s.Require().NoError(config.LoadConfig())
	s.Require().NoError(config.SetupLogging())
	s.config = config

	s.keys = []string{"fee", "fi", "fo", "fum"}

	s.tracker, err = acomm.NewTracker(filepath.Join(dir, "tracker.sock"), nil, nil, 5*time.Second)
	s.Require().NoError(err)

	s.KV, err = New(s.config, s.tracker)
	s.Require().NoError(err)
}

func (s *KVS) SetupTest() {
	for _, str := range s.keys {
		key := s.KVPrefix + "/" + str
		s.Require().NoError(s.Suite.KV.Set(key, key))
		s.Require().NoError(s.Suite.KV.Set(key+"-dir/"+str+"1", key+"-dir/"+str+"1"))
		s.Require().NoError(s.Suite.KV.Set(key+"-dir/"+str+"2", key+"-dir/"+str+"2"))
	}
}

func (s *KVS) TestConfig() {
	addr, err := s.config.Address()
	s.Require().NoError(err)
	s.Require().Equal(s.KVURL, addr)
}

func (s *KVS) TestHandleKVDown() {
	s.KVCmd.Process.Signal(syscall.SIGSTOP)
	provider, err := New(s.config, s.tracker)
	s.Require().NoError(err)

	s.True(provider.kvDown())

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-get",
		Args: GetArgs{Key: "some-non-existent-key"},
	})
	s.Require().NoError(err)
	resp, url, err := provider.get(req)
	s.Nil(resp)
	s.Nil(url)
	s.NotNil(err)
	s.Equal(errorKVDown, err)

	type temporary interface {
		Temporary() bool
	}
	temp, ok := err.(temporary)
	s.True(ok, "error should implement Temporary interface")
	s.True(temp.Temporary())

	s.KVCmd.Process.Signal(syscall.SIGCONT)

	kvDown := true
	for range [5]struct{}{} {
		kvDown = kvDown && provider.kvDown()
		if !kvDown {
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}
	s.False(kvDown)
	resp, url, err = provider.get(req)
	s.Nil(resp)
	s.Nil(url)
	s.NotNil(err)
	s.NotEqual(errorKVDown, err)
}
