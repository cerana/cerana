package zfs

import (
	"errors"
	"net/url"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/mistifyio/gozfs"
	"github.com/mistifyio/mistify/acomm"
)

// Send returns information about a dataset.
func (z *ZFS) Send(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CommonArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	ds, err := gozfs.GetDataset(args.Name)
	if err != nil {
		return nil, nil, err
	}

	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	addr, err := z.tracker.NewStreamUnix(z.config.StreamDir("zfs-send"), reader)
	if err != nil {
		return nil, nil, err
	}

	go func() {
		defer func() {
			if err := writer.Close(); err != nil {
				logrus.WithField("error", err).Error("failed to close snapshot stream writer")
			}
		}()
		if err := ds.Send(writer); err != nil {
			logrus.WithField("error", err).Error("failed to send snapshot")
		}
	}()

	return nil, addr, err
}
