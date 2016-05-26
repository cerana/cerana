package kv

import (
	"time"

	"github.com/cerana/cerana/acomm"
)

func (s *KVS) TestLockKnownBad() {
	tests := []struct {
		name string
		key  string
		ttl  time.Duration
		err  string
	}{
		{name: "no key", err: "missing arg: key"},
		{name: "no ttl", key: "foo", err: "missing arg: ttl"},
	}

	for _, test := range tests {
		args := LockArgs{
			Key: test.key,
			TTL: test.ttl,
		}

		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "kv-lock",
			Args: args,
		})
		s.Require().NoError(err, test.name)

		res, streamURL, err := s.KV.lock(req)
		s.EqualError(err, test.err, test.name)
		s.Nil(streamURL)
		s.Nil(res)
	}
}

// a port of pkg/kv/kv_test.go's TestLock
func (s *KVS) TestLock() {
	lockReq, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-lock",
		Args: LockArgs{
			Key: s.PrefixKey("some-lock"),
			TTL: 1 * time.Second,
		},
	})
	s.Require().NoError(err)

	// acquire lock
	res, streamURL, err := s.KV.lock(lockReq)
	s.Require().NoError(err, "should be able to acquire lock")
	s.Require().Nil(streamURL)
	s.Require().NotNil(res)

	lock := res.(Cookie)

	res, streamURL, err = s.KV.lock(lockReq)
	s.Require().Error(err, "should not be able to acquire an acquired lock")
	s.Require().Nil(streamURL)
	s.Require().Nil(res)

	// unlocking
	unlockReq, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-unlock",
		Args: lock,
	})
	res, streamURL, err = s.KV.unlock(unlockReq)
	s.Require().NoError(err, "unlocking should not fail")
	s.Require().Nil(streamURL)
	s.Require().Nil(res)

	res, streamURL, err = s.KV.unlock(unlockReq)
	s.Require().Error(err, "unlocking lost lock should fail")
	s.Require().Nil(streamURL)
	s.Require().Nil(res)

	res, streamURL, err = s.KV.lock(lockReq)
	s.Require().NoError(err, "acquiring an unlocked lock should pass")
	s.Require().Nil(streamURL)
	s.Require().NotNil(res)

	lock = res.(Cookie)

	renewReq, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-renew",
		Args: lock,
	})
	for i := 0; i < 5; i++ {
		res, streamURL, err = s.KV.renew(renewReq)
		s.Require().NoError(err, "renewing a lock should pass")
		s.Require().Nil(streamURL)
		s.Require().Nil(res)
		time.Sleep(1 * time.Second)
	}

	time.Sleep(3 * time.Second)
	res, streamURL, err = s.KV.renew(renewReq)
	s.Require().Error(err, "renewing an expired lock should fail")
	s.Require().Nil(streamURL)
	s.Require().Nil(res)

	// consul's default lock-delay
	// see lock-delay at https://www.consul.io/docs/internals/sessions.html
	time.Sleep(15 * time.Second)

	res, streamURL, err = s.KV.lock(lockReq)
	s.Require().NoError(err, "should be able to acquire previously expired lock")
	s.Require().Nil(streamURL)
	s.Require().NotNil(res)
}
