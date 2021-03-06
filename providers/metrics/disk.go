package metrics

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/shirou/gopsutil/disk"
)

// DiskResult is the result for the Disk handler.
type DiskResult struct {
	IO         map[string]disk.IOCountersStat `json:"io"`
	Partitions []disk.PartitionStat           `json:"partitions"`
	Usage      []*disk.UsageStat              `json:"usage"`
}

// Disk returns information about the disk partitions, io, and usage.
func (m *Metrics) Disk(req *acomm.Request) (interface{}, *url.URL, error) {
	io, err := disk.IOCounters()
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}

	partitions, err := disk.Partitions(true)
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}

	usage := make([]*disk.UsageStat, 0, len(partitions))
	for _, partition := range partitions {
		u, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			return nil, nil, errors.Wrap(err)
		}
		if u.InodesTotal != 0 {
			usage = append(usage, u)
		}
	}

	return &DiskResult{IO: io, Partitions: partitions, Usage: usage}, nil, nil
}
