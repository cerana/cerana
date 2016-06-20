package etcd

import (
	"testing"
	"time"

	"github.com/cerana/cerana/internal/tests/common"
	"github.com/stretchr/testify/suite"
)

func TestEtcdDetail(t *testing.T) {
	suite.Run(t, &ETCD{})
}

type ETCD struct {
	common.Suite
}

func (s *ETCD) SetupSuite() {
	s.KVCmdMaker = common.EtcdMaker
	s.Suite.SetupSuite()
}

func (s *ETCD) TestLockEmbeddedValue() {
	Junk := s.KVPrefix + "/lock-key-garbage"
	s.Require().NoError(s.KV.Set(Junk, "locked=foobar"))

	True := s.KVPrefix + "/lock-key-true"
	s.Require().NoError(s.KV.Set(True, "locked=true"))

	False := s.KVPrefix + "/lock-key-false"
	s.Require().NoError(s.KV.Set(False, "locked=false"))

	lock, err := s.KV.Lock(True, 1*time.Second)
	s.Error(err)
	s.EqualError(err, "101: Compare failed ([locked=false != locked=true]) [6]")
	s.Nil(lock)

	lock, err = s.KV.Lock(False, 1*time.Second)
	s.NoError(err)
	s.NotNil(lock)

	lock, err = s.KV.Lock(Junk, 1*time.Second)
	s.Error(err)
	s.Nil(lock)
}
