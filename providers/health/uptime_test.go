package health_test

import (
	"fmt"
	"time"

	"github.com/cerana/cerana/acomm"
	healthp "github.com/cerana/cerana/providers/health"
	"github.com/cerana/cerana/providers/systemd"
	"github.com/pborman/uuid"
)

func (s *health) TestUptime() {
	goodStatus := s.addService()

	tests := []struct {
		name        string
		minUptime   time.Duration
		expectedErr string
	}{
		{"", time.Second, "missing arg: name"},
		{"foobar", time.Second, "No such file or directory"},
		{goodStatus.Name, time.Second, ""},
		{goodStatus.Name, time.Minute, ""},
		{goodStatus.Name, time.Hour, "uptime less than expected"},
	}

	for _, test := range tests {
		args := healthp.UptimeArgs{
			Name:      test.name,
			MinUptime: test.minUptime,
		}
		desc := fmt.Sprintf("%+v", args)

		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task:         "health-uptime",
			ResponseHook: s.responseHook,
			Args:         args,
		})
		s.Require().NoError(err, desc)

		resp, stream, err := s.health.Uptime(req)
		s.Nil(resp, desc)
		s.Nil(stream, desc)
		if test.expectedErr == "" {
			s.Nil(err, desc)
		} else {
			s.EqualError(err, test.expectedErr, desc)
		}
	}
}

func (s *health) addService() systemd.UnitStatus {
	name := uuid.New()
	s.systemd.ManualCreate(systemd.CreateArgs{
		Name: name,
	}, true)
	return s.systemd.Data.Statuses[name]
}
