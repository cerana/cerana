package service_test

import (
	"fmt"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/service"
	"github.com/cerana/cerana/providers/systemd"
	"github.com/pborman/uuid"
)

func (s *Provider) TestRemove() {
	id := uuid.New()
	bundleID := uint64(1)
	name := fmt.Sprintf("%d:%s.service", bundleID, id)
	s.systemd.ManualCreate(systemd.CreateArgs{Name: name}, true)

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "service-remove",
		Args: service.RemoveArgs{
			ID:       id,
			BundleID: bundleID,
		},
	})

	result, streamURL, err := s.provider.Remove(req)
	s.Nil(result)
	s.Nil(streamURL)
	s.NoError(err)

	s.Nil(s.systemd.ManualGet(name))
}
