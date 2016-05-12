package kv

import (
	"encoding/json"
	"net/http"

	"github.com/cerana/cerana/acomm"
)

func (s *KVS) getIndex(url string, key string) uint64 {
	resp, err := http.Get(url + "/v1/kv/" + key)
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()

	m := []map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&m)
	s.Require().NoError(err)
	return uint64(m[0]["ModifyIndex"].(float64))
}

func (s *KVS) TestRemove() {
	tests := []struct {
		name  string
		key   string
		index uint64
		err   string
	}{
		{"no key", "", 0, "missing arg: key"},
		{"bad index", s.KVPrefix + "/" + s.keys[0], 2, "failed to delete atomically"},
		{"good index", s.KVPrefix + "/" + s.keys[0], 1 /*will be changed*/, ""},
	}

	for _, test := range tests {
		args := RemoveArgs{
			Key:   test.key,
			Index: test.index,
		}
		if test.index == 1 {
			args.Index = s.getIndex(s.KVURL, test.key)
		}

		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "kv-remove",
			Args: args,
		})
		s.Require().NoError(err, test.name)

		res, streamURL, err := s.KV.remove(req)
		if test.err != "" {
			s.EqualError(err, test.err, test.name)
			continue
		}

		if !s.Nil(err, test.name) {
			continue
		}

		s.Empty(streamURL, test.name)

		if !s.Nil(res, test.name) {
			continue
		}

		_, err = s.Suite.KV.Get(test.key)
		s.True(s.Suite.KV.IsKeyNotFound(err))
	}
}

func (s *KVS) TestUpdate() {
	tests := []struct {
		name  string
		key   string
		value string
		index uint64
		err   string
	}{
		{"no key", "", "", 0, "missing arg: key"},
		{"no value", "foo", "", 0, "missing arg: value"},
		{"create existing key", s.keys[0], "foo", 0, "CAS failed"},
		{"valid create", "cas-update-test", "cas-update-test", 0, ""},
		{"bad index", "cas-update-test", "cas-update-test-bad", 2, "CAS failed"},
		{"good index", "cas-update-test", "cas-update-test-good", 1 /*will be changed*/, ""},
	}

	lastIndex := uint64(0)
	for _, test := range tests {
		args := UpdateArgs{
			Key:   test.key,
			Value: test.value,
			Index: test.index,
		}
		if test.index == 1 {
			args.Index = lastIndex
		}

		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "kv-update",
			Args: args,
		})
		s.Require().NoError(err, test.name)

		res, streamURL, err := s.KV.update(req)
		if test.err != "" {
			s.EqualError(err, test.err, test.name)
			continue
		}

		if !s.Nil(err, test.name) {
			continue
		}

		s.Empty(streamURL, test.name)

		if !s.NotNil(res, test.name) {
			continue
		}

		val, err := s.Suite.KV.Get(test.key)
		s.NoError(err)
		s.Equal(test.value, string(val.Data))

		lastIndex = res.(UpdateReturn).Index
	}
}
