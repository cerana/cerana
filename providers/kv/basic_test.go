package kv

import (
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/kv"
)

func (s *KVS) TestDelete() {
	tests := []struct {
		name string
		key  string
		err  string
	}{
		{"no key", "", "missing arg: key"},
		{"valid key", s.KVPrefix + "/" + s.keys[0], ""},
	}

	for _, test := range tests {
		args := GetArgs{
			Key: test.key,
		}
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "kv-delete",
			Args: args,
		})
		s.Require().NoError(err, test.name)

		if test.err == "" {
			_, err = s.Suite.KV.Get(test.key)
			s.NoError(err)
		}

		res, streamURL, err := s.KV.delete(req)
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

func (s *KVS) TestGet() {
	tests := []struct {
		name string
		key  string
		data []byte
		err  string
	}{
		{"no key", "", nil, "missing arg: key"},
		{"non-existent-key", "some-non-existent-key", nil, "key not found"},
		{"dir-not-key", s.KVPrefix + "/" + s.keys[0] + "-dir/", nil, "key not found"},
		{"valid key", s.KVPrefix + "/" + s.keys[0], nil, ""},
	}

	for _, test := range tests {
		args := GetArgs{Key: test.key}
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "kv-get",
			Args: args,
		})
		s.Require().NoError(err, test.name)

		res, streamURL, err := s.KV.get(req)
		s.Empty(streamURL, test.name)
		if test.err != "" {
			s.EqualError(err, test.err, test.name)
			continue
		}

		if !s.Nil(err, test.name) {
			continue
		}
		if !s.NotNil(res, test.name) {
			continue
		}

		result, ok := res.([]byte)
		if !s.True(ok, test.name) {
			continue
		}
		if !s.NotNil(result) {
			continue
		}

		s.Equal(test.key, string(result))
	}
}

func (s *KVS) TestGetAll() {
	tests := []struct {
		name string
		key  string
		len  int
		err  string
	}{
		{"no key", "", 0, "missing arg: key"},
		{"valid key", s.KVPrefix + "/" + s.keys[0] + "-dir", 2, ""},
		{"all keys", s.KVPrefix, len(s.keys) * 3, ""},
	}

	for _, test := range tests {
		args := GetArgs{
			Key: test.key,
		}
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "kv-getAll",
			Args: args,
		})
		s.Require().NoError(err, test.name)

		res, streamURL, err := s.KV.getAll(req)
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

		values, ok := res.(map[string]kv.Value)
		s.True(ok)
		s.Len(values, test.len)
	}
}

func (s *KVS) TestKeys() {
	tests := []struct {
		name string
		key  string
		len  int
		err  string
	}{
		{"no key", "", 0, "missing arg: key"},
		{"valid key", s.KVPrefix + "/" + s.keys[0] + "-dir", 2, ""},
		{"top level keys", s.KVPrefix, len(s.keys) * 2, ""},
	}

	for _, test := range tests {
		args := GetArgs{
			Key: test.key,
		}
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "kv-keys",
			Args: args,
		})
		s.Require().NoError(err, test.name)

		res, streamURL, err := s.KV.keys(req)
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

		keys, ok := res.([]string)
		s.True(ok)
		s.Len(keys, test.len)
	}
}

func (s *KVS) TestSet() {
	tests := []struct {
		name string
		key  string
		data string
		err  string
	}{
		{"no key", "", "", "missing arg: key"},
		{"no data", "foo", "", "missing arg: data"},
		{"valid key", s.keys[0], "foo", ""},
	}

	for _, test := range tests {
		args := SetArgs{
			Key:  test.key,
			Data: test.data,
		}
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "kv-set",
			Args: args,
		})
		s.Require().NoError(err, test.name)

		res, streamURL, err := s.KV.set(req)
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

		val, err := s.Suite.KV.Get(test.key)
		s.NoError(err)
		s.Equal(test.data, string(val.Data))
	}
}
