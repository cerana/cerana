// Package common contains common utilities and suites to be used in other tests
package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/cerana/cerana/pkg/kv"
	_ "github.com/cerana/cerana/pkg/kv/consul" // register consul with pkg/kv
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ConsulMaker will create an exec.Cmd to run consul with the given paramaters
func ConsulMaker(port uint16, dir, prefix string) *exec.Cmd {
	b, err := json.Marshal(map[string]interface{}{
		"ports": map[string]interface{}{
			"dns":      port + 1,
			"http":     port + 2,
			"rpc":      port + 3,
			"serf_lan": port + 4,
			"serf_wan": port + 5,
			"server":   port + 6,
		},
		"session_ttl_min": "1s",
	})
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(dir+"/config.json", b, 0444)
	if err != nil {
		panic(err)
	}

	return exec.Command("consul",
		"agent",
		"-server",
		"-bootstrap-expect", "1",
		"-config-file", dir+"/config.json",
		"-data-dir", dir,
		"-bind", "127.0.0.1",
		"-http-port", strconv.Itoa(int(port)),
	)
}

// EtcdMaker will create an exec.Cmd to run etcd with the given paramaters
func EtcdMaker(port uint16, dir, prefix string) *exec.Cmd {
	clientURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	peerURL := fmt.Sprintf("http://127.0.0.1:%d", port+1)
	return exec.Command("etcd",
		"-name", prefix,
		"-data-dir", dir,
		"-initial-cluster-state", "new",
		"-initial-cluster-token", prefix,
		"-initial-cluster", prefix+"="+peerURL,
		"-initial-advertise-peer-urls", peerURL,
		"-listen-peer-urls", peerURL,
		"-listen-client-urls", clientURL,
		"-advertise-client-urls", clientURL,
	)
}

// Suite sets up a general test suite with setup/teardown.
type Suite struct {
	suite.Suite
	KVDir      string
	KVPrefix   string
	KVPort     uint16
	KVURL      string
	KV         kv.KV
	KVCmd      *exec.Cmd
	KVCmdMaker func(uint16, string, string) *exec.Cmd
	TestPrefix string
}

// SetupSuite runs a new kv instance.
func (s *Suite) SetupSuite() {
	if s.TestPrefix == "" {
		s.TestPrefix = "lochness-test"
	}

	s.KVDir, _ = ioutil.TempDir("", s.TestPrefix+"-"+uuid.New())

	if s.KVPort == 0 {
		s.KVPort = uint16(1024 + rand.Intn(65535-1024))
	}

	if s.KVCmdMaker == nil {
		s.KVCmdMaker = ConsulMaker
	}
	s.KVCmd = s.KVCmdMaker(s.KVPort, s.KVDir, s.TestPrefix)

	if testing.Verbose() {
		s.KVCmd.Stdout = os.Stdout
		s.KVCmd.Stderr = os.Stderr
	}
	s.Require().NoError(s.KVCmd.Start())
	time.Sleep(2500 * time.Millisecond) // Wait for test kv to be ready

	var err error
	for i := 0; i < 10; i++ {
		s.KV, err = kv.New("http://127.0.0.1:" + strconv.Itoa(int(s.KVPort)))
		if err == nil {
			break
		}
		err = s.KV.Ping()
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond) // Wait for test kv to be ready
	}
	if s.KV == nil {
		panic(err)
	}

	s.KVPrefix = "lochness"
	s.KVURL = "http://127.0.0.1:" + strconv.Itoa(int(s.KVPort))
}

// SetupTest prepares anything needed per test.
func (s *Suite) SetupTest() {
}

// TearDownTest cleans the kv instance.
func (s *Suite) TearDownTest() {
	s.Require().NoError(s.KV.Delete(s.KVPrefix, true))
}

// TearDownSuite stops the kv instance and removes all data.
func (s *Suite) TearDownSuite() {
	// Stop the test kv process
	s.Require().NoError(s.KVCmd.Process.Kill())
	s.Require().Error(s.KVCmd.Wait())

	// Remove the test kv data directory
	_ = os.RemoveAll(s.KVDir)
}

// PrefixKey generates a kv key using the set prefix
func (s *Suite) PrefixKey(key string) string {
	return filepath.Join(s.KVPrefix, key)
}

// DoRequest is a convenience method for making an http request and doing basic handling of the response.
func (s *Suite) DoRequest(method, url string, expectedRespCode int, postBodyStruct interface{}, respBody interface{}) *http.Response {
	var postBody io.Reader
	if postBodyStruct != nil {
		bodyBytes, _ := json.Marshal(postBodyStruct)
		postBody = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, url, postBody)
	if postBody != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	s.NoError(err)
	correctResponse := s.Equal(expectedRespCode, resp.StatusCode)
	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	s.NoError(err)

	if correctResponse {
		s.NoError(json.Unmarshal(body, respBody))
	} else {
		s.T().Log(string(body))
	}
	return resp
}
