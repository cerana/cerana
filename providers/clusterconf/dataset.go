package clusterconf

import "net"

// Dataset is configuration for a dataset.
type Dataset struct {
	ID                string          `json:"id"`
	Parent            string          `json:"parent"`
	ParentSameMachine bool            `json:"parentSameMachine"`
	ReadOnly          bool            `json:"readOnly"`
	NFS               bool            `json:"nfs"`
	Redundancy        int             `json:"redundancy"`
	Quota             int             `json:"quota"`
	Nodes             map[net.IP]bool `json:"nodes"`
}
