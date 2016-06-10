package health

import (
	"errors"
	"io/ioutil"
	"net"
	"net/url"
	"regexp"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/logrusx"
)

// TCPResponseArgs are arguments for TCPResponse health checks.
type TCPResponseArgs struct {
	Address string `json:"address"`
	Body    []byte `json:"body"`
	Regexp  string `json:"regexp"`
}

// TCPResponse makes a TCP request to the specified address and checks the
// response for a match to a specified regex.
func (h *Health) TCPResponse(req *acomm.Request) (interface{}, *url.URL, error) {
	var args TCPResponseArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Address == "" {
		return nil, nil, errors.New("missing arg: address")
	}
	if args.Regexp == "" {
		return nil, nil, errors.New("missing arg: regexp")
	}
	re, err := regexp.Compile(args.Regexp)
	if err != nil {
		return nil, nil, err
	}

	conn, err := net.DialTimeout("tcp", args.Address, h.config.RequestTimeout())
	if err != nil {
		return nil, nil, err
	}
	defer logrusx.LogReturnedErr(conn.Close, nil, "failed to close tcp conn")
	if len(args.Body) > 0 {
		if _, err = conn.Write(args.Body); err != nil {
			return nil, nil, err
		}
		if err = conn.(*net.TCPConn).CloseWrite(); err != nil {
			return nil, nil, err
		}
	}

	tcpResp, err := ioutil.ReadAll(conn)
	if err != nil {
		return nil, nil, err
	}

	if !re.Match(tcpResp) {
		return nil, nil, errors.New("response did not match")
	}
	return nil, nil, nil
}
