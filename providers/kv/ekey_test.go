package kv

import (
	"time"

	"github.com/cerana/cerana/acomm"
)

func (s *KVS) get(key string) string {
	v, err := s.KV.kv.Get(key)
	s.Require().NoError(err)
	return string(v.Data)
}

func (s *KVS) TestEKeyKnownBad() {
	tests := []struct {
		name  string
		key   string
		value string
		ttl   time.Duration
		err   string
	}{
		{name: "no key", err: "missing arg: key"},
		{name: "no value", key: "foo", err: "missing arg: value"},
		{name: "no ttl", key: "foo", value: "foo", err: "missing arg: ttl"},
	}

	for _, test := range tests {
		args := EphemeralSetArgs{
			Key:   test.key,
			Value: test.value,
			TTL:   test.ttl,
		}

		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "kv-ehpemeral-set",
			Args: args,
		})
		s.Require().NoError(err, test.name)

		res, streamURL, err := s.KV.eset(req)
		s.EqualError(err, test.err, test.name)
		s.Nil(streamURL)
		s.Nil(res)
	}
}

func (s *KVS) TestEKey() {
	key := s.PrefixKey("test-ekey")
	ekeyReq, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-ephemeral-set",
		Args: EphemeralSetArgs{
			Key:   key,
			Value: "1",
			TTL:   1 * time.Second,
		},
	})
	s.Require().NoError(err)

	// create ekey
	res, streamURL, err := s.KV.eset(ekeyReq)
	s.Require().NoError(err, "should be able to set ekey")
	s.Require().Nil(streamURL)
	s.Require().Nil(res)
	s.Require().Equal("1", s.get(key))

	ekeyReq, err = acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-ephemeral-set",
		Args: EphemeralSetArgs{
			Key:   key,
			Value: "2",
			TTL:   1 * time.Second,
		},
	})
	s.Require().NoError(err)

	// set ekey to something new
	res, streamURL, err = s.KV.eset(ekeyReq)
	s.Require().NoError(err, "should be able to reset ekey")
	s.Require().Nil(streamURL)
	s.Require().Nil(res)
	s.Require().Equal("2", s.get(key))

	// renew
	for i := 0; i < 5; i++ {
		res, streamURL, err = s.KV.eset(ekeyReq)
		s.Require().NoError(err, "(re)setting an ekey should also renew")
		s.Require().Nil(streamURL)
		s.Require().Nil(res)
		time.Sleep(1 * time.Second)
	}
	s.Require().Equal("2", s.get(key))

	// ekey should expire
	time.Sleep(3 * time.Second)
	time.Sleep(15 * time.Second) // lock delay
	_, err = s.KV.kv.Get(key)
	s.Require().Error(err, "ekey should expire after ttl has passed")

	// should be able to create new key after expiration of old
	res, streamURL, err = s.KV.eset(ekeyReq)
	s.Require().NoError(err, "should be able to create new ekey, after expiration of old")
	s.Require().Nil(streamURL)
	s.Require().Nil(res)
	s.Require().Equal("2", s.get(key))
}

func (s *KVS) TestEDestroy() {
	// check expected fail conditions
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-ehpemeral-destroy",
		Args: EphemeralDestroyArgs{},
	})
	s.Require().NoError(err)

	res, streamURL, err := s.KV.edestroy(req)
	s.EqualError(err, "missing arg: key")
	s.Nil(streamURL)
	s.Nil(res)

	req, err = acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-ehpemeral-destroy",
		Args: EphemeralDestroyArgs{
			Key: "non-existent-key",
		},
	})
	s.Require().NoError(err)

	res, streamURL, err = s.KV.edestroy(req)
	s.EqualError(err, "unknown ephemeral key")
	s.Nil(streamURL)
	s.Nil(res)

	// test good conditions
	key := s.PrefixKey("test-ekey")
	ekeyReq, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-ephemeral-set",
		Args: EphemeralSetArgs{
			Key:   key,
			Value: "1",
			TTL:   1 * time.Second,
		},
	})
	s.Require().NoError(err)

	// ensure ekey created
	res, streamURL, err = s.KV.eset(ekeyReq)
	s.Require().NoError(err, "should be able to set ekey")
	s.Require().Nil(streamURL)
	s.Require().Nil(res)
	s.Require().Equal("1", s.get(key))

	req, err = acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-ehpemeral-destroy",
		Args: EphemeralDestroyArgs{
			Key: key,
		},
	})
	s.Require().NoError(err)

	res, streamURL, err = s.KV.edestroy(req)
	s.NoError(err)
	s.Nil(streamURL)
	s.Nil(res)

	_, err = s.KV.kv.Get(key)
	s.Require().Error(err, "ekey should not exist")
}
