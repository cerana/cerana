package zfs

import (
	"fmt"
	"strings"

	"github.com/mistifyio/mistify/provider"
)

// ZFS is a provider of zfs functionality.
type ZFS struct{}

// CommonArgs are arguments that apply to all handlers.
type CommonArgs struct {
	Name string `json:"name"` // Name of dataset
}

// RegisterTasks registers all of ZFS's task handlers with the server.
func (z *ZFS) RegisterTasks(server *provider.Server) {
	server.RegisterTask("zfs-exists", z.Exists)
	server.RegisterTask("zfs-destroy", z.Destroy)
	server.RegisterTask("zfs-create", z.Create)
	server.RegisterTask("zfs-holds", z.Holds)
	server.RegisterTask("zfs-snapshot", z.Snapshot)
}

// fixPropertyTypes attempts to convert the underlying data types in a property
// map that came from JSON to what zfs expects.
func fixPropertyTypesFromJSON(properties map[string]interface{}) error {
	for key, origValue := range properties {
		if strings.Contains(key, ":") {
			properties[key] = fmt.Sprintf("%v", origValue)
			continue
		}
		if origValue, ok := origValue.(float64); ok {
			newValue := uint64(origValue)
			if float64(newValue) != origValue {
				return fmt.Errorf("property %s must be a uint64: %v", key, origValue)
			}
			properties[key] = uint64(origValue)
		}
	}
	return nil
}
