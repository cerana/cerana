package clusterconf_test

import (
	"fmt"
	"time"

	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/spf13/pflag"
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
				s.Contains(err.Error(), fmt.Sprintf("invalid %s", ttlType), desc)
			}
		}
	}
}

func (s *clusterConf) TestLoadConfig() {
	datasetTTL := s.config.DatasetTTL()
	bundleTTL := s.config.DatasetTTL()
	nodeTTL := s.config.DatasetTTL()
	defer s.viper.Set("dataset_ttl", datasetTTL.String())
	defer s.viper.Set("bundle_ttl", bundleTTL.String())
	defer s.viper.Set("node_ttl", nodeTTL.String())

	v := s.coordinator.NewProviderViper()
	flagset := pflag.NewFlagSet("clusterconfLoadConfig", pflag.PanicOnError)
	// Note these explicit flags can be removed with issue #149
	flagset.DurationP("dataset_ttl", "d", time.Minute, "ttl for dataset usage heartbeats")
	flagset.DurationP("bundle_ttl", "b", time.Minute, "ttl for bundle usage heartbeats")
	flagset.DurationP("node_ttl", "o", time.Minute, "ttl for node heartbeats")
	config := clusterconf.NewConfig(flagset, v)
	s.NoError(flagset.Parse([]string{
		"--dataset_ttl", "123s",
		"--bundle_ttl", "456s",
		"--node_ttl", "789s",
	}))
	if !s.NoError(config.LoadConfig()) {
		return
	}
	s.Equal(123*time.Second, config.DatasetTTL())
	s.Equal(456*time.Second, config.BundleTTL())
	s.Equal(789*time.Second, config.NodeTTL())

}
