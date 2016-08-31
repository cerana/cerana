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
	"strings"
	"sync"
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
func ConsulMaker(port uint16, dir, prefix string) ([]*exec.Cmd, []string) {
	nodes := 3
	cmds := make([]*exec.Cmd, nodes)
	addrs := make([]string, nodes)

	for i := range addrs {
		addrs[i] = fmt.Sprintf("127.0.0.1%d", i)
	}

	for i := range cmds {
		name := prefix + strconv.Itoa(i)
		dir := filepath.Join(dir, strconv.Itoa(i))
		err := os.Mkdir(dir, 0700)
		if err != nil {
			panic(err)
		}
		conf := dir + "/config.json"

		b, err := json.Marshal(map[string]interface{}{
			"session_ttl_min": "1s",
		})
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(conf, b, 0444)
		if err != nil {
			panic(err)
		}

		args := []string{
			"agent", "-server",
			"-bind", addrs[i],
			"-client", addrs[i],
			"-config-file", conf,
			"-data-dir", dir,
			"-http-port", strconv.Itoa(int(port)),
			"-node", name,
		}
		if i == 0 {
			args = append(args, "-bootstrap-expect", strconv.Itoa(nodes))
		} else {
			args = append(args, "-join", addrs[0])
		}
		cmds[i] = exec.Command("consul", args...)
	}
	for i := range addrs {
		addrs[i] = fmt.Sprintf("http://%s:%d", addrs[i], port)
	}
	return cmds, addrs
}

// EtcdMaker will create an exec.Cmd to run etcd with the given paramaters
func EtcdMaker(port uint16, dir, prefix string) ([]*exec.Cmd, []string) {
	nodes := 3
	cmds := make([]*exec.Cmd, nodes)
	peerURLs := make([]string, nodes)
	clientURLs := make([]string, nodes)
	names := make([]string, nodes)
	clusterList := make([]string, nodes)

	for i := range peerURLs {
		name := fmt.Sprintf("%s%d", prefix, i)
		peerURL := fmt.Sprintf("http://127.0.0.1%d", i)
		clusterList[i] = fmt.Sprintf("%s=%s", name, peerURL)
		names[i] = name
		peerURLs[i] = peerURL
	}
	cluster := strings.Join(clusterList, ",")

	for i := range cmds {
		clientURL := fmt.Sprintf("http://127.0.0.1%d:%d", i, port)
		cmd := exec.Command("etcd",
			"-name", names[i],
			"-data-dir", fmt.Sprintf("%s%d", dir, i),
			"-initial-cluster-state", "new",
			"-initial-cluster-token", prefix,
			"-initial-cluster", cluster,
			"-initial-advertise-peer-urls", peerURLs[i],
			"-listen-peer-urls", peerURLs[i],
			"-listen-client-urls", clientURL,
			"-advertise-client-urls", clientURL,
		)
		clientURLs[i] = clientURL
		cmds[i] = cmd
	}
	return cmds, clientURLs
}

// Suite sets up a general test suite with setup/teardown.
type Suite struct {
	suite.Suite
	KVDir      string
	KVPrefix   string
	KVPort     uint16
	KVURLs     []string
	KV         kv.KV
	KVCmds     []*exec.Cmd
	KVCmdMaker func(uint16, string, string) ([]*exec.Cmd, []string)
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
	s.KVCmds, s.KVURLs = s.KVCmdMaker(s.KVPort, s.KVDir, s.TestPrefix)

	for _, cmd := range s.KVCmds {
		if testing.Verbose() {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		s.Require().NoError(cmd.Start())
	}
	time.Sleep(5000 * time.Millisecond) // Wait for test kv to be ready

	var err error
	for i := 0; i < 10; i++ {
		s.KV, err = kv.New("http://127.0.0.10:" + strconv.Itoa(int(s.KVPort)))
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
	// Stop the test kv processes
	wg := sync.WaitGroup{}
	wg.Add(len(s.KVCmds))
	for _, cmd := range s.KVCmds {
		cmd := cmd
		go func() {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			wg.Done()
		}()
	}
	wg.Wait()

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
