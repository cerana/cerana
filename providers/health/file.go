package health

import (
	"errors"
	"net/url"
	"os"

	"github.com/cerana/cerana/acomm"
)

// FileArgs are arguments for the File health check.
type FileArgs struct {
	Path     string      `json:"path"`
	NotExist bool        `json:"notExist"`
	Mode     os.FileMode `json:"mode"`
	MinSize  int64       `json:"minSize"`
	MaxSize  int64       `json:"maxSize"`
}

// File checks one or more attributes against supplied constraints.
func (h *Health) File(req *acomm.Request) (interface{}, *url.URL, error) {
	var args FileArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Path == "" {
		return nil, nil, errors.New("missing arg: path")
	}

	fileInfo, err := os.Stat(args.Path)
	if err != nil {
		if args.NotExist && os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	} else if args.NotExist {
		return nil, nil, errors.New("file exists")
	}

	if args.Mode != 0 && (args.Mode&fileInfo.Mode() != args.Mode) {
		return nil, nil, errors.New("unexpected mode")
	}
	if fileInfo.Size() < args.MinSize {
		return nil, nil, errors.New("size below min")
	}
	if args.MaxSize > 0 && fileInfo.Size() > args.MaxSize {
		return nil, nil, errors.New("size above max")
	}

	return nil, nil, nil
}
