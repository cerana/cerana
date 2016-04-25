package clusterconf_test

import (
	"fmt"
	"time"
)

func (s *clusterConf) TestConfigTTLs() {
	ttlTypes := []struct {
		key string
		fn  func() time.Duration
	}{
		{"dataset_ttl", s.config.DatasetTTL},
		{"bundle_ttl", s.config.BundleTTL},
		{"node_ttl", s.config.NodeTTL},
	}

	for _, ttlType := range ttlTypes {
		s.Equal(time.Minute, ttlType.fn(), ttlType.key)
	}
}

func (s *clusterConf) TestValidate() {
	datasetTTL := s.config.DatasetTTL()
	bundleTTL := s.config.DatasetTTL()
	nodeTTL := s.config.DatasetTTL()
	defer s.viper.Set("dataset_ttl", datasetTTL)
	defer s.viper.Set("bundle_ttl", bundleTTL)
	defer s.viper.Set("node_ttl", nodeTTL)

	ttlTypes := []string{
		"dataset_ttl",
		"bundle_ttl",
		"node_ttl",
	}

	tests := []struct {
		duration time.Duration
		valid    bool
	}{
		{0, false},
		{-1, false},
		{-1 * time.Second, false},
		{time.Second, true},
		{time.Minute, true},
		{time.Hour, true},
	}

	for _, ttlType := range ttlTypes {
		for _, test := range tests {
			desc := fmt.Sprintf("%s : %s", ttlType, test.duration)
			s.viper.Set(ttlType, test.duration)
			err := s.config.Validate()
			if test.valid {
				s.NoError(err, desc)
			} else {
				s.EqualError(err, fmt.Sprintf("invalid %s", ttlType), desc)
			}
		}
	}
}
