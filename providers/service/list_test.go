package service_test

import (
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/service"
	"github.com/cerana/cerana/providers/systemd"
	"github.com/pborman/uuid"
)

func (s *Provider) TestList() {
	s.systemd.ManualCreate(systemd.CreateArgs{Name: "foobar"}, true)
	s.systemd.ManualCreate(systemd.CreateArgs{Name: "1:" + uuid.New() + ".service"}, true)
	s.systemd.ManualCreate(systemd.CreateArgs{Name: "2:" + uuid.New() + ".service"}, true)

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:         "service-list",
		ResponseHook: s.responseHook,
	})
	s.Require().NoError(err)

	listResult, streamURL, err := s.provider.List(req)
	s.Nil(streamURL)
	s.NoError(err)
	if !s.NotNil(listResult) {
		return
	}
	result, ok := listResult.(service.ListResult)
	if !s.True(ok) {
		return
	}
	s.Len(result.Services, 2)
}
