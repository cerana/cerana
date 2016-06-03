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
	defer s.viper.Set("dataset_ttl", datasetTTL.String())
	defer s.viper.Set("bundle_ttl", bundleTTL.String())
	defer s.viper.Set("node_ttl", nodeTTL.String())

	ttlTypes := []string{
		"dataset_ttl",
		"bundle_ttl",
		"node_ttl",
	}

	tests := []struct {
		duration string
		valid    bool
	}{
		{"0", false},
		{"1", false},
		{"-1s", false},
		{"1s", true},
		{"1m", true},
		{"1h", true},
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
