package namespace

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/logrusx"
)

// UserArgs are arguments for SetUser.
type UserArgs struct {
	PID  uint64  `json:"pid"`
	UIDs []IDMap `json:"uid"`
	GIDs []IDMap `json:"gid"`
}

// IDMap is a map of id in container to id on host and length of a range.
type IDMap struct {
	ID     uint64 `json:"id"`
	HostID uint64 `json:"hostID"`
	Length uint64 `json:"Length"`
}

// SetUser sets the user and group id mapping for a process.
func (n *Namespace) SetUser(req *acomm.Request) (interface{}, *url.URL, error) {
	var args UserArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	uidMapPath := fmt.Sprintf("/proc/%d/uid_map", args.PID)
	if err := writeIDMapFile(uidMapPath, args.UIDs); err != nil {
		return nil, nil, err
	}

	gidMapPath := fmt.Sprintf("/proc/%d/gid_map", args.PID)
	if err := writeIDMapFile(gidMapPath, args.GIDs); err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func writeIDMapFile(path string, idMaps []IDMap) error {
	mapFile, err := os.OpenFile(path, os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer logrusx.LogReturnedErr(mapFile.Close, logrus.Fields{"path": path}, "failed to close map file")

	content := make([]string, len(idMaps))
	for i, idMap := range idMaps {
		content[i] = fmt.Sprintf("%d %d %d", idMap.ID, idMap.HostID, idMap.Length)
	}

	if _, err := fmt.Fprintln(mapFile, strings.Join(content, "\n")); err != nil {
		return err
	}

	return nil
}
