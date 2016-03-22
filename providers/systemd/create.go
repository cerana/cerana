package systemd

import (
	"errors"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/coreos/go-systemd/unit"
	"github.com/mistifyio/mistify/acomm"
)

// CreateArgs are arguments for the Create handler.
type CreateArgs struct {
	Name        string             `json:"name"`
	UnitOptions []*unit.UnitOption `json:"unit-options"`
}

// Create creates a new unit file.
func (s *Systemd) Create(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CreateArgs
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

	if _, err := os.Stat(unitFilePath); err == nil {
		return nil, nil, errors.New("unit file already exists")
	}

	unitFileContents, err := ioutil.ReadAll(unit.Serialize(args.UnitOptions))
	// TODO: Sort out file permissions
	return nil, nil, ioutil.WriteFile(unitFilePath, unitFileContents, os.ModePerm)
}
