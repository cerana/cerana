package systemd

import (
	"errors"
	"net/url"
	"os"

	"github.com/cerana/cerana/acomm"
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
		return nil, nil, errors.New("missing arg: name")
	}

	unitFilePath, err := s.config.UnitFilePath(args.Name)
	if err != nil {
		return nil, nil, err
	}

	if err := os.Remove(unitFilePath); err != nil && !os.IsNotExist(err) {
		return nil, nil, err
	}
	return nil, nil, nil
}
