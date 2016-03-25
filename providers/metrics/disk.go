package metrics

import (
	"net/url"

	"github.com/mistifyio/mistify/acomm"
	"github.com/shirou/gopsutil/disk"
)

type DiskResult struct {
	IO         map[string]disk.DiskIOCountersStat `json:"io"`
	Partitions []disk.DiskPartitionStat           `json:"partitions"`
	Usage      []*disk.DiskUsageStat              `json:"usage"`
}

func (m *Metrics) Disk(req *acomm.Request) (interface{}, *url.URL, error) {
	io, err := disk.DiskIOCounters()
	if err != nil {
		return nil, nil, err
	}

	partitions, err := disk.DiskPartitions(true)
	if err != nil {
		return nil, nil, err
	}

	usage := make([]*disk.DiskUsageStat, 0, len(partitions))
	for _, partition := range partitions {
		u, err := disk.DiskUsage(partition.Mountpoint)
		if err != nil {
			return nil, nil, err
		}
		if u.InodesTotal != 0 {
			usage = append(usage, u)
		}
	}

	return &DiskResult{IO: io, Partitions: partitions, Usage: usage}, nil, nil
}
