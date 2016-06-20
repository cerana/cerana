package kv

import (
	"encoding/json"
	"io"
	"math/rand"
	"net"
	"strconv"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/kv"
)

func mkEvent(eType kv.EventType, key string, index uint64) Event {
	event := Event{}
	event.Key = key
	event.Type = eType
	event.Index = index
	switch eType {
	case kv.Create:
		event.Data = []byte(key)
	case kv.Update:
		event.Data = []byte(key + key)
	case kv.Delete:
	}

	return event
}

func (s *KVS) setupWatch(prefix string, index uint64) (Cookie, net.Conn) {
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-watch",
		Args: WatchArgs{
			Prefix: prefix,
			Index:  index,
		},
	})
	s.Require().NoError(err)

	resp, url, err := s.KV.watch(req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(url)

	conn, err := net.Dial("unix", url.RequestURI())
	s.Require().NoError(err)

	return resp.(Cookie), conn
}

func (s *KVS) TestWatch() {
	type setter func(string) Event

	events := []Event{}
	prefix := s.KVPrefix + "/watch-test/"
	tests := []struct {
		name   string
		prefix string
		index  uint64
		event  Event
		setter setter
		err    string
	}{
		{name: "no prefix", err: "missing arg: prefix"},
		{name: "no index", prefix: prefix, err: "missing arg: index"},
		{name: "create", prefix: prefix, setter: func(key string) Event {
			err := s.Suite.KV.Set(key, key)
			if err != nil {
				s.T().Fatal(err)
			}
			return mkEvent(kv.Create, key, s.getIndex(s.KVURL, key))
		}},
		{name: "update", prefix: prefix, setter: func(key string) Event {
			err := s.Suite.KV.Set(key, key+key)
			if err != nil {
				s.T().Fatal(err)
			}
			return mkEvent(kv.Update, key, s.getIndex(s.KVURL, key))
		}},
		{name: "delete", prefix: prefix, setter: func(key string) Event {
			index := s.getIndex(s.KVURL, key)
			err := s.Suite.KV.Delete(key, false)
			if err != nil {
				s.T().Fatal(err)
			}
			return mkEvent(kv.Delete, key, index)
		}},
	}

	index := s.getIndex(s.KVURL, s.KVPrefix+"/"+s.keys[0])
	_, conn := s.setupWatch(prefix, index)

	key := strconv.Itoa(rand.Int())
	for _, test := range tests {
		if test.setter == nil {
			continue
		}

		event := test.setter(test.prefix + key)
		events = append(events, event)
	}

	dec := json.NewDecoder(conn)
	event := Event{}
	for i := range events {
		s.NoError(dec.Decode(&event))
		s.Equal(events[i], event)
	}
}

func (s *KVS) TestStopMissingCookie() {
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-stop",
		Args: Cookie{},
	})
	s.Require().NoError(err)

	resp, url, err := s.KV.stop(req)
	s.Require().Nil(resp)
	s.Require().Nil(url)
	s.Require().Equal("missing arg: cookie", err.Error())
}

func (s *KVS) TestStopNoActivity() {
	index := s.getIndex(s.KVURL, s.KVPrefix+"/"+s.keys[0])
	cookie, conn := s.setupWatch(s.KVPrefix, index)

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-stop",
		Args: cookie,
	})

	dec := json.NewDecoder(conn)
	resp, url, err := s.KV.stop(req)
	s.Require().NoError(err)
	s.Require().Nil(resp)
	s.Require().Nil(url)

	event := Event{}
	err = dec.Decode(&event)
	s.Equal(io.EOF, err)
}

func (s *KVS) TestStopWithActivity() {
	index := s.getIndex(s.KVURL, s.KVPrefix+"/"+s.keys[0])
	cookie, conn := s.setupWatch(s.KVPrefix, index)

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-stop",
		Args: cookie,
	})

	dec := json.NewDecoder(conn)
	resp, url, err := s.KV.stop(req)
	s.Require().NoError(err)
	s.Require().Nil(resp)
	s.Require().Nil(url)

	s.Require().NoError(s.Suite.KV.Set(s.KVPrefix+"/"+s.keys[0], "foo"))

	event := Event{}
	err = dec.Decode(&event)
	s.Equal(io.EOF, err)
}
