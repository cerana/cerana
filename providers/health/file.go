package health

import (
	"net/url"
	"os"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
)

// FileArgs are arguments for the File health check.
type FileArgs struct {
	Path     string      `json:"path"`
	NotExist bool        `json:"notExist"`
	Mode     os.FileMode `json:"mode"`
	MinSize  int64       `json:"minSize"`
	MaxSize  int64       `json:"maxSize"`
}

// File checks one or more attributes against supplied attributes.
func (h *Health) File(req *acomm.Request) (interface{}, *url.URL, error) {
	var args FileArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Path == "" {
		return nil, nil, errors.Newv("missing arg: path", map[string]interface{}{"args": args})
	}

	fileInfo, err := os.Stat(args.Path)
	if err != nil {
		if args.NotExist && os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, errors.Wrapv(err, map[string]interface{}{"path": args.Path})
	} else if args.NotExist {
		return nil, nil, errors.Newv("file exists", map[string]interface{}{"path": args.Path})
	}

	if args.Mode != 0 && (args.Mode&fileInfo.Mode() != args.Mode) {
		return nil, nil, errors.Newv("unexpected mode", map[string]interface{}{"mode": fileInfo.Mode(), "expectedMode": args.Mode})
	}
	if fileInfo.Size() < args.MinSize {
		return nil, nil, errors.Newv("size below min", map[string]interface{}{"size": fileInfo.Size(), "minSize": args.MinSize})
	}
	if args.MaxSize > 0 && fileInfo.Size() > args.MaxSize {
		return nil, nil, errors.Newv("size above max", map[string]interface{}{"size": fileInfo.Size(), "maxSize": args.MaxSize})
	}

	return nil, nil, nil
}
