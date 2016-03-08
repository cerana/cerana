package zfs

import "github.com/mistifyio/mistify/provider"

// ZFS is a provider of zfs functionality.
type ZFS struct{}

// CommonArgs are arguments that apply to all handlers.
type CommonArgs struct {
	Name string `json:"name"` // Name of dataset
}

// RegisterTasks registers all of ZFS's task handlers with the server.
func (z *ZFS) RegisterTasks(server *provider.Server) {
	server.RegisterTask("zfs-exists", z.Exists)
}
