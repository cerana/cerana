package clusterconf

import "net"

// Bundle is configuration for a bundle of services.
type Bundle struct {
	ID         int                       `json:"id"`
	Datasets   map[string]*BundleDataset `json:"datasets"`
	Services   map[string]*BundleService `json:"services"`
	Redundancy int                       `json:"redundancy"`
	Nodes      map[string]net.IP         `json:"nodes"`
	Ports      map[int]*BundlePort       `json:"ports"`
}

// BundleDataset is configuration for a dataset associated with a bundle.
type BundleDataset struct {
	Name      string `json:"name"`
	DatasetID string `json:"datasetID"`
	Type      int    `json:"type"` // TODO: Decide on type for this. Iota?
	Quota     int    `json:"type"`
}

// BundleService is configuration overrides for a service of a bundle and
// associated datasets.
type BundleService struct {
	*Service
	Datasets map[string]*ServiceDataset `json:"datasets"`
}

// ServiceDataset is configuration for mounting a dataset for a bundle service.
type ServiceDataset struct {
	Name       string `json:"name"`
	MountPoint string `json:"mountPoint"`
	ReadOnly   bool   `json:"readOnly"`
}

// BundlePort is configuration for a port associated with a bundle.
type BundlePort struct {
	Port             int      `json:"port"`
	Public           bool     `json:"public"`
	ConnectedBundles []string `json:"connectedBundles"`
	ExternalPort     int      `json:"externalPort"`
}
