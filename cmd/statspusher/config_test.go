package main

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/suite"
)

type StatsPusher struct {
	suite.Suite
}

func TestStatsPusher(t *testing.T) {
	suite.Run(t, new(StatsPusher))
}

func (s *StatsPusher) TestCanonicalFlagName() {
	tests := []struct {
		input  string
		output string
	}{
		{"foo bar", "fooBar"},
		{"foo	bar", "fooBar"},
		{"foo.bar", "fooBar"},
		{"foo_bar", "fooBar"},
		{"foo-bar", "fooBar"},
		{"foo.-_bar", "fooBar"},
		{"Foo_bar", "fooBar"},
		{"foo_bar.baz bang", "fooBarBazBang"},
		{"Foo_bar", "fooBar"},
		{"foo_BAR", "fooBAR"},
		{"foo_cpu", "fooCPU"},
		{"foo_ttl", "fooTTL"},
		{"foo_cpu", "fooCPU"},
		{"foo_url", "fooURL"},
		{"foo_ip", "fooIP"},
		{"foo_id", "fooID"},
		{"FOO_bar", "FOOBar"},
		{"Cpu_bar", "cpuBar"},
		{"id_bar", "idBar"},
		{"CPU_bar", "cpuBar"},
	}

	for _, test := range tests {
		s.Equal(test.output, string(canonicalFlagName(pflag.CommandLine, test.input)), test.input)
	}
}
