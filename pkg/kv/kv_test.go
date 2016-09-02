package kv_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/cerana/cerana/internal/tests/common"
	"github.com/cerana/cerana/pkg/kv"
	consul "github.com/cerana/cerana/pkg/kv/consul"
	etcd "github.com/cerana/cerana/pkg/kv/etcd"
	"github.com/stretchr/testify/suite"
)

func TestKV(t *testing.T) {
	suite.Run(t, &KVSuite{})
}

type KVSuite struct {
	keys []string
	common.Suite
}

func (s *KVSuite) SetupSuite() {
	switch os.Getenv("KV") {
	case "", "consul":
	case "etcd":
		s.KVCmdMaker = common.EtcdMaker
	default:
		panic("unknown KV specified in environment")
	}
	s.Suite.SetupSuite()
	s.keys = []string{s.KVPrefix + "/fee", s.KVPrefix + "/fi", s.KVPrefix + "/fo", s.KVPrefix + "/fum"}
}

func (s *KVSuite) SetupTest() {
	for _, str := range s.keys {
		s.Require().NoError(s.KV.Set(str, str))
		s.Require().NoError(s.KV.Set(str+"-dir/"+str+"1", str+"-dir/"+str+"1"))
		s.Require().NoError(s.KV.Set(str+"-dir/"+str+"2", str+"-dir/"+str+"2"))
	}
}

func (s *KVSuite) TestRegister() {
	s.Require().Panics(func() {
		kv.Register("consul", func(string) (kv.KV, error) { return nil, nil })
	}, "should panic when trying register a duplicate")
}

func (s *KVSuite) TestEtcdNew() {
	tests := []struct {
		addr string
		err  bool
	}{
		{"%zz", true},
		{"", false},
		{"etcd://", false},
		{"http://", false},
	}
	for _, test := range tests {
		_, err := etcd.New(test.addr)
		if test.err != (err != nil) {
			want := "no error"
			if test.err {
				want = "an error"
			}

			s.Fail(fmt.Sprintf("error mismatch want: %s, got: %v", want, err))
		}
	}
}

func (s *KVSuite) TestConsulNew() {
	tests := []struct {
		addr string
		err  bool
	}{
		{"%zz", true},
		{"", false},
		{"consul://", false},
		{"http://", false},
	}
	for _, test := range tests {
		_, err := consul.New(test.addr)
		if test.err != (err != nil) {
			want := "no error"
			if test.err {
				want = "an error"
			}

			s.Fail(fmt.Sprintf("error mismatch want: %s, got: %v", want, err))
		}
	}
}

func (s *KVSuite) TestNew() {
	c, _ := consul.New("")
	h := c
	e, _ := etcd.New("")
	if os.Getenv("KV") == "etcd" {
		h = e
	}
	tests := []struct {
		addr string
		err  bool
		kv   kv.KV
	}{
		{"%zz", true, nil},
		{"", true, nil},
		{"kvite://", true, nil},
		{"etcd://", true, e},
		{fmt.Sprintf("consul://127.0.0.1:%d", s.KVPort), false, c},
		{fmt.Sprintf("http://127.0.0.1:%d", s.KVPort), false, h},
	}
	for _, test := range tests {
		_, err := kv.New(test.addr)
		if test.err != (err != nil) {
			want := "no error"
			if test.err {
				want = "an error"
			}
			s.Fail(fmt.Sprintf("error mismatch want: %s, got: %v", want, err))
		}
	}
}

func (s *KVSuite) TestPing() {
	s.Require().NoError(s.KV.Ping())
}

func (s *KVSuite) TestIsKeyNotFound() {
	s.Require().Panics(func() { get(s.KVPort, "cerana/non-existent-key") })
	_, err := s.KV.Get("cerana/non-existent-key")
	s.Require().True(s.KV.IsKeyNotFound(err))
}

func getConsul(port uint16, key string) string {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/kv/%s", port, key))
	if err != nil {
		panic(err)
	}
	defer func() { _ = resp.Body.Close() }()

	var m []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		panic(err)
	}

	b, err := base64.StdEncoding.DecodeString(m[0]["Value"].(string))
	if err != nil {
		panic(err)
	}
	return string(b)
}

func get(port uint16, key string) string {
	switch os.Getenv("KV") {
	case "etcd":
		panic("Not Implemented Yet")
	default:
		return getConsul(port, key)
	}
}

func (s *KVSuite) TestSet() {
	s.Require().Error(s.KV.Set("/some-file", "some-value"))
	for _, str := range []string{"FEE", "FI", "FO", "FUM"} {
		key := s.KVPrefix + "/" + str
		s.Require().NoError(s.KV.Set(key, str))
		s.Require().Equal(str, get(s.KVPort, key))
		s.Require().NoError(s.KV.Set(key, str+str))
		s.Require().Equal(str+str, get(s.KVPort, key))
	}
}

func (s *KVSuite) TestGet() {
	_, err := s.KV.Get("some-non-existent-file")
	s.Require().Error(err, "Get non-existent key should have failed")
	for _, str := range s.keys {
		s.Require().NoError(s.KV.Set(str, str))
		val, err := s.KV.Get(str)
		s.Require().NoError(err)
		s.Require().Equal(val.Data, []byte(str))

		s.Require().NoError(s.KV.Set(str, str+str))
		val, err = s.KV.Get(str)
		s.Require().NoError(err)
		s.Require().Equal(val.Data, []byte(str+str))
	}
}

func (s *KVSuite) TestGetAll() {
	m, err := s.KV.GetAll("some-prefix")
	s.Require().NoError(err)
	s.Require().Empty(m)

	m, err = s.KV.GetAll(s.KVPrefix)
	s.Require().NoError(err)
	s.Require().Len(m, len(s.keys)*3)
	for k, v := range m {
		s.Require().Equal(k, string(v.Data))
	}
}

func (s *KVSuite) TestKeys() {
	m, err := s.KV.Keys("some-prefix")
	s.Require().NoError(err)
	s.Require().Empty(m)

	m, err = s.KV.Keys(s.KVPrefix)
	s.Require().NoError(err)
	s.Require().Len(m, 8)

	s.Require().NoError(s.KV.Set(s.KVPrefix+"/foo/bar", "bar"))
	m, err = s.KV.Keys(s.KVPrefix)
	s.Require().NoError(err)
	s.Require().Len(m, 9)
}

func (s *KVSuite) TestDelete() {
	delKey := s.keys[0]
	delDir := s.keys[1] + "-dir"

	s.Require().NoError(s.KV.Delete(delKey, false))
	m, err := s.KV.GetAll(s.KVPrefix)
	s.Require().NoError(err)
	s.Require().Len(m, len(s.keys)*3-1)

	s.Require().NoError(s.KV.Delete(delDir, true))
	m, err = s.KV.GetAll(s.KVPrefix)
	s.Require().NoError(err)
	s.Require().Len(m, len(s.keys)*3-3)
}

func (s *KVSuite) TestUpdate() {
	_, err := s.KV.Update("some-key", kv.Value{})
	s.Require().Error(err)

	_, err = s.KV.Get("cerana/some-key")
	s.Require().True(s.KV.IsKeyNotFound(err))

	idx, err := s.KV.Update("cerana/some-key", kv.Value{Data: []byte("1")})
	s.Require().NoError(err)
	s.Require().True(idx > 0)

	_, err = s.KV.Update("cerana/some-key", kv.Value{Data: []byte("2")})
	s.Require().Error(err)
	_, err = s.KV.Update("cerana/some-key", kv.Value{Data: []byte("2"), Index: idx - 1})
	s.Require().Error(err)

	idx2, err := s.KV.Update("cerana/some-key", kv.Value{Data: []byte("2"), Index: idx})
	s.Require().NoError(err)
	s.Require().True(idx2 > idx)
}

func (s *KVSuite) TestRemove() {
	s.Require().NoError(s.KV.Remove("some-random-key", 0))

	idx, err := s.KV.Update("cerana/some-key", kv.Value{Data: []byte("1")})
	s.Require().NoError(err)
	s.Require().True(idx > 0)

	s.Require().Error(s.KV.Remove("cerana/some-key", idx-1))
	v, err := s.KV.Get("cerana/some-key")
	s.Require().NoError(err)
	s.Require().True(v.Index == idx)
	s.Require().Equal([]byte("1"), v.Data)

	s.Require().Error(s.KV.Remove("cerana/some-key", idx+1))
	v, err = s.KV.Get("cerana/some-key")
	s.Require().NoError(err)
	s.Require().True(v.Index == idx)
	s.Require().Equal([]byte("1"), v.Data)

	s.Require().NoError(s.KV.Remove("cerana/some-key", idx))
	_, err = s.KV.Get("cerana/some-key")
	s.Require().Error(err)
	s.Require().True(s.KV.IsKeyNotFound(err))
}

func (s *KVSuite) TestWatch() {
	index, err := s.KV.Update("cerana/some-key", kv.Value{Data: []byte("1")})
	s.Require().NoError(err)
	s.Require().True(index > 0)

	stop := make(chan struct{})
	events, errors, err := s.KV.Watch("cerana/", index, stop)
	s.Require().NoError(err)
	s.Require().NotNil(events)
	s.Require().NotNil(errors)
	select {
	case event := <-events:
		s.Require().FailNow("unexpected event: %v", event)
	case err := <-errors:
		s.Require().FailNow("unexpected error: %v", err)
	default:
	}

	tests := []struct {
		name  string
		data  []byte
		eType kv.EventType
		index uint64
	}{
		{name: "create", eType: kv.Create},
		{name: "update", eType: kv.Update},
		{name: "delete", eType: kv.Delete},
	}

	key := "cerana/watcher-test-key"
	index = 0
	for i := range tests {
		t := &tests[i]
		t.data = []byte(t.name)
		if tests[i].name != "delete" {
			var err error
			index, err = s.KV.Update(key, kv.Value{Data: t.data, Index: index})
			s.Require().NoError(err)
		} else {
			t.data = nil
			s.Require().NoError(s.KV.Delete(key, false))
		}
		t.index = index

		event := s.getEvent(events, errors)
		exp := kv.Event{
			Key:  key,
			Type: t.eType,
			Value: kv.Value{
				Data:  t.data,
				Index: t.index,
			},
		}
		s.Require().Equal(exp, event)
	}

	close(stop)
	s.Require().NoError(s.KV.Set(key, ""))

	for event := range events {
		s.Require().FailNow("unexpected event", "%v", event)
	}
	for err := range errors {
		s.Require().FailNow("unexpected error", "%v", err)
	}
}

func (s *KVSuite) getEvent(events chan kv.Event, errors chan error) kv.Event {
	select {
	case event := <-events:
		return event
	case err := <-errors:
		s.Require().FailNow("unexpected error: %v", err)
	case <-time.After(100 * time.Millisecond):
		s.Require().FailNow("timeout waiting for an expected event")
	}
	panic("should not get here")
}

func (s *KVSuite) makeEKey(key string) kv.EphemeralKey {
	ekey, err := s.KV.EphemeralKey(key, 1*time.Second)
	s.Require().NoError(err)

	s.Require().NoError(ekey.Set("init"))
	s.Require().Equal("init", get(s.KVPort, key))
	return ekey
}

func (s *KVSuite) TestEphemeralSet() {
	key := "cerana/ekey-set"
	ekey := s.makeEKey(key)

	s.Require().NoError(ekey.Set("value"))
	v, err := s.KV.Get(key)
	s.Require().NoError(err)
	s.Require().Equal("value", string(v.Data))
}

func (s *KVSuite) TestEphemeralDestroy() {
	key := "cerana/ekey-destroy"
	ekey := s.makeEKey(key)

	s.Require().NoError(ekey.Renew())
	s.Require().NoError(ekey.Destroy())
	_, err := s.KV.Get(key)
	s.Require().Error(err)
}

func (s *KVSuite) TestEphemeralRenew() {
	s.T().Parallel()

	key := "cerana/ekey-renew"
	ekey := s.makeEKey(key)

	for i := 0; i < 5; i++ {
		s.Require().NoError(ekey.Renew())
		time.Sleep(1 * time.Second)
		_, err := s.KV.Get(key)
		s.Require().NoError(err)
	}

	err := ekey.Renew()
	time.Sleep(3 * time.Second)
	_, err = s.KV.Get(key)
	s.Require().True(s.KV.IsKeyNotFound(err))
}

// this test has been ported to providers/kv, any change here should probably be reflected there too
func (s *KVSuite) TestLock() {
	s.T().Parallel()

	key := "cerana/lock"

	// acquire lock
	lock, err := s.KV.Lock(key, 1*time.Second)
	s.Require().NoError(err, "should be able to acquire lock")

	_, err = s.KV.Lock(key, 1*time.Second)
	s.Require().Error(err, "should not be able to acquire an acquired lock")

	s.Require().NoError(lock.Unlock(), "unlocking should not fail")
	s.Require().Error(lock.Unlock(), "unlocking lost lock should fail")

	_, err = s.KV.Lock(key, 1*time.Second)
	s.Require().NoError(err, "acquiring an unlocked lock should pass")

	for i := 0; i < 5; i++ {
		s.Require().NoError(lock.Renew())
		time.Sleep(1 * time.Second)
	}

	time.Sleep(3 * time.Second)
	s.Require().Error(lock.Renew(), "renewing an expired lock should fail")

	// consul's default lock-delay
	// see lock-delay at https://www.consul.io/docs/internals/sessions.html
	time.Sleep(15 * time.Second)

	lock, err = s.KV.Lock(key, 1*time.Second)
	s.Require().NoError(err, "should be able to acquire previously expired lock")

}
