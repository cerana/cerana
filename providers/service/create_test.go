package service_test

import (
	"fmt"
	"strings"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/service"
	"github.com/pborman/uuid"
)

func (s *Provider) TestCreate() {
	tests := []struct {
		id          string
		bundleID    uint64
		description string
		exec        []string
		env         map[string]string
		err         string
	}{
		{"", 219, "working service", []string{"foo", "bar"}, map[string]string{"foo": "bar"}, "missing arg: id"},
		{uuid.New(), 0, "working service", []string{"foo", "bar"}, map[string]string{"foo": "bar"}, "missing arg: bundleID"},
		{uuid.New(), 219, "", []string{"foo", "bar"}, map[string]string{"foo": "bar"}, ""},
		{uuid.New(), 219, "working service", nil, map[string]string{"foo": "bar"}, "missing arg: exec"},
		{uuid.New(), 219, "working service", []string{}, map[string]string{"foo": "bar"}, "missing arg: exec"},
		{uuid.New(), 219, "working service", []string{"foo", "bar"}, map[string]string{}, ""},
		{uuid.New(), 219, "working service", []string{"foo", "bar"}, map[string]string{"foo": "bar"}, ""},
		{uuid.New(), 219, "working service", []string{"foo", "bar"}, map[string]string{"_CERANA_foo": "bar"}, ""},
	}

	for _, test := range tests {
		args := &service.CreateArgs{
			ID:          test.id,
			BundleID:    test.bundleID,
			Description: test.description,
			Exec:        test.exec,
			Env:         test.env,
		}
		desc := fmt.Sprintf("%+v", args)
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "service-create",
			Args: args,
		})
		s.Require().NoError(err, desc)

		result, streamURL, err := s.provider.Create(req)
		s.Nil(streamURL, desc)
		if test.err != "" {
			s.EqualError(err, test.err, desc)
			s.Nil(result, desc)
		} else {
			s.NoError(err, desc)
			if !s.NotNil(result, desc) {
				continue
			}
			getResult, ok := result.(service.GetResult)
			if !s.True(ok, desc) {
				continue
			}
			s.Equal(test.id, getResult.Service.ID, desc)
			s.Equal(test.bundleID, getResult.Service.BundleID, desc)
			s.Equal(test.description, getResult.Service.Description, desc)
			s.Equal(test.exec, getResult.Service.Exec, desc)
			for key, val := range test.env {
				if strings.HasPrefix(key, "_CERANA_") {
					_, ok := getResult.Service.Env[key]
					s.False(ok, desc)
				} else {
					s.Equal(val, getResult.Service.Env[key], desc)
				}
			}
		}
	}
}
