package service_test

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/service"
	"github.com/cerana/cerana/providers/systemd"
	"github.com/coreos/go-systemd/unit"
	"github.com/pborman/uuid"
)

func (s *Provider) TestGet() {
	id := uuid.New()
	bundleID := uint64(219)
	name := fmt.Sprintf("%d:%s.service", bundleID, id)
	desc := "foobar"
	execStart := []string{"foo", "bar"}
	user := uint64(123)
	group := uint64(456)
	env := map[string]string{"foo": "bar"}
	envStr := "foo=bar"
	s.systemd.ManualCreate(systemd.CreateArgs{
		Name: name,
		UnitOptions: []*unit.UnitOption{
			{Section: "Unit", Name: "Description", Value: desc},
			{Section: "Service", Name: "ExecStart", Value: strings.Join(execStart, " ")},
			{Section: "Service", Name: "User", Value: strconv.FormatUint(user, 10)},
			{Section: "Service", Name: "Group", Value: strconv.FormatUint(group, 10)},
			{Section: "Service", Name: "Environment", Value: envStr},
		},
	}, true)
	svc := s.systemd.ManualGet(name)

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:         "service-get",
		ResponseHook: s.responseHook,
		Args: service.GetArgs{
			ID:       id,
			BundleID: bundleID,
		},
	})
	s.Require().NoError(err)
	getResult, streamURL, err := s.provider.Get(req)
	s.Nil(streamURL)
	s.NoError(err)
	if !s.NotNil(getResult) {
		return
	}
	result, ok := getResult.(service.GetResult)
	if !s.True(ok) {
		return
	}

	s.Equal(id, result.Service.ID)
	s.Equal(bundleID, result.Service.BundleID)
	s.Equal(desc, result.Service.Description)
	s.Equal(svc.Uptime, result.Service.Uptime)
	s.Equal(execStart, result.Service.Cmd)
	s.Equal(user, result.Service.UID)
	s.Equal(group, result.Service.GID)
	s.Equal(env, result.Service.Env)
}
