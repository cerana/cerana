package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify/acomm"
	flags "github.com/spf13/pflag"
)

const (
	argSep = "="
)

func main() {
	log.SetLevel(log.FatalLevel)

	var coordinator, responseAddr, taskName string
	var taskArgs []string
	flags.StringVarP(&coordinator, "coordinator_url", "c", "", "url of the coordinator")
	flags.StringVarP(&taskName, "task", "t", "", "task to run")
	flags.StringSliceVarP(&taskArgs, "request_arg", "a", []string{}, fmt.Sprintf("task specific argument the form 'key%svalue'. can be set multiple times", argSep))
	flags.StringVarP(&responseAddr, "response_addr", "r", ":4080", "address for response http handler to listen on")
	flags.Parse()

	args, err := parseTaskArgs(taskArgs)
	dieOnError(err)

	result, stream, respErr, err := startResponseServer(responseAddr)
	dieOnError(err)

	dieOnError(makeRequest(coordinator, taskName, responseAddr, args))

	select {
	case err := <-respErr:
		dieOnError(err)
	case result := <-result:
		j, _ := json.Marshal(result)
		fmt.Println(string(j))
	case stream := <-stream:
		dieOnError(acomm.Stream(os.Stdout, stream))
	}
}

func dieOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// argmap is an arbitrarilly nested map
type argmap map[string]interface{}

func (am argmap) set(keys []string, value interface{}) error {
	if len(keys) == 0 {
		fmt.Println("no more keys and returning")
		return nil
	}

	key := keys[0]

	if len(keys) == 1 {
		am[key] = value
		return nil
	}

	var m argmap
	mi, ok := am[key]
	if !ok {
		m = make(argmap)
		am[key] = m
	} else {
		m, ok = mi.(argmap)
		if !ok {
			return fmt.Errorf("intermediate nested key %s defined and not a map", key)
		}
	}

	return m.set(keys[1:], value)
}

func parseTaskArgs(taskArgs []string) (map[string]interface{}, error) {
	out := make(argmap)
	for _, in := range taskArgs {
		parts := strings.Split(in, argSep)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid request arg: '%s'", in)
		}

		valueS := strings.Join(parts[1:], argSep)
		var value interface{}
		if arg, err := strconv.ParseInt(valueS, 10, 64); err == nil {
			value = arg
		} else if arg, err := strconv.ParseBool(valueS); err == nil {
			value = arg
		} else {
			value = valueS
		}

		keys := strings.Split(parts[0], ".")
		if err := out.set(keys, value); err != nil {
			return nil, err
		}

	}
	return out, nil
}

func startResponseServer(addr string) (chan interface{}, chan *url.URL, chan error, error) {
	result := make(chan interface{}, 1)
	errChan := make(chan error, 1)
	stream := make(chan *url.URL, 1)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			errChan <- err
			return
		}

		resp := &acomm.Response{}
		if err := json.Unmarshal(body, resp); err != nil {
			errChan <- err
			return
		}

		ack, _ := json.Marshal(&acomm.Response{})
		_, _ = w.Write(ack)

		if resp.Error != nil {
			errChan <- resp.Error
			return
		}

		if resp.StreamURL != nil {
			stream <- resp.StreamURL
		} else {
			result <- resp.Result
		}
	})

	runErr := make(chan error)
	running := time.NewTimer(time.Second)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			runErr <- err
		}
	}()

	var err error
	select {
	case <-running.C:
	case err = <-runErr:
	}

	return result, stream, errChan, err
}

func makeRequest(coordinator, taskName, responseAddr string, taskArgs map[string]interface{}) error {
	coordinatorURL, err := url.ParseRequestURI(coordinator)
	if err != nil {
		return errors.New("invalid coordinator url")
	}

	responseHook := fmt.Sprintf("http://%s/", responseAddr)
	req, err := acomm.NewRequest(taskName, responseHook, "", taskArgs, nil, nil)
	if err != nil {
		return err
	}

	return acomm.Send(coordinatorURL, req)
}
