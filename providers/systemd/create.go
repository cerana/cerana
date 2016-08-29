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

// Create creates a new unit file.
func (s *Systemd) Create(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CreateArgs
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

	if !args.Overwrite {
		if _, err = os.Stat(unitFilePath); err == nil {
			return nil, nil, errors.Newv("unit file already exists", map[string]interface{}{"path": unitFilePath})
		}
	}

	unitFileContents, err := ioutil.ReadAll(unit.Serialize(args.UnitOptions))
	if err != nil {
		return nil, nil, errors.Wrapv(err, map[string]interface{}{"unitOptions": args.UnitOptions})
	}
	// TODO: Sort out file permissions
	if err := ioutil.WriteFile(unitFilePath, unitFileContents, os.ModePerm); err != nil {
		return nil, nil, errors.Wrapv(err, map[string]interface{}{
			"path":     unitFilePath,
			"contents": string(unitFileContents),
			"perms":    os.ModePerm,
		})
	}

	return nil, nil, errors.Wrap(s.dconn.Reload())
}
