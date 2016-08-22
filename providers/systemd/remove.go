package systemd

import (
	"net/url"
	"os"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
)

// RemoveArgs are arguments for the Remove handler.
type RemoveArgs struct {
	Name string `json:"name"`
}

// Remove deletes a unit file.
func (s *Systemd) Remove(req *acomm.Request) (interface{}, *url.URL, error) {
	var args RemoveArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.Newv("missing arg: name", map[string]interface{}{"args": args})
	}

	unitFilePath, err := s.config.UnitFilePath(args.Name)
	if err != nil {
		return nil, nil, err
	}

	if err := os.Remove(unitFilePath); err != nil && !os.IsNotExist(err) {
		return nil, nil, errors.Wrapv(err, map[string]interface{}{"path": unitFilePath})
	}
	return nil, nil, nil
}
