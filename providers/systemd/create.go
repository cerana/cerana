package systemd

import (
	"io/ioutil"
	"net/url"
	"os"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/coreos/go-systemd/unit"
)

// CreateArgs are arguments for the Create handler.
type CreateArgs struct {
	Name        string             `json:"name"`
	UnitOptions []*unit.UnitOption `json:"unit-options"`
	Overwrite   bool               `json:"overwrite"`
}

// CreateResult is the result of a create action.
type CreateResult struct {
	UnitModified bool `json:"modified"`
}

// Create creates or overwrites a unit file.
func (s *Systemd) Create(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CreateArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.Newv("missing arg: name", map[string]interface{}{"args": args})
	}

	unitFileContents, err := ioutil.ReadAll(unit.Serialize(args.UnitOptions))
	if err != nil {
		return nil, nil, errors.Wrapv(err, map[string]interface{}{"unitOptions": args.UnitOptions})
	}

	unitFilePath, err := s.config.UnitFilePath(args.Name)
	if err != nil {
		return nil, nil, err
	}

	if _, err = os.Stat(unitFilePath); err == nil {
		if !args.Overwrite {
			return nil, nil, errors.Newv("unit file already exists", map[string]interface{}{"path": unitFilePath})
		}

		// check if modifications exist to avoid unnecessary work
		origUnitFileContents, err := ioutil.ReadFile(unitFilePath)
		if err != nil {
			return nil, nil, errors.Wrapv(err, map[string]interface{}{"path": unitFilePath})
		}

		if string(unitFileContents) == string(origUnitFileContents) {
			return CreateResult{}, nil, nil
		}
	}

	// TODO: Sort out file permissions
	if err := ioutil.WriteFile(unitFilePath, unitFileContents, os.ModePerm); err != nil {
		return nil, nil, errors.Wrapv(err, map[string]interface{}{
			"path":     unitFilePath,
			"contents": string(unitFileContents),
			"perms":    os.ModePerm,
		})
	}

	if err := errors.Wrap(s.dconn.Reload()); err != nil {
		return nil, nil, err
	}
	return CreateResult{true}, nil, nil
}
