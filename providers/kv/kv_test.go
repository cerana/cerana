package kv

import (
	"io/ioutil"
	"testing"

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
	config *Config
	KV     *KV
	keys   []string
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

	s.KV, err = New(s.config)
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
