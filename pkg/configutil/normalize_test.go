package configutil_test

import (
	"strings"
	"testing"

	"github.com/cerana/cerana/pkg/configutil"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/suite"
)

type ConfigUtil struct {
	suite.Suite
}

func TestConfigUtil(t *testing.T) {
	suite.Run(t, new(ConfigUtil))
}

func (s *ConfigUtil) TestNormalizeFunc() {
	type testStruct struct {
		input  string
		output string
	}

	tests := []testStruct{
		{"foo", "foo"},
		{"Foo", "foo"},
		{"FOO", "FOO"},
		{"fooBar", "fooBar"},
		{"FooBar", "fooBar"},
		{"fooBBar", "fooBBar"},
		{" fooBar ", "fooBar"},
	}

	sets := []struct {
		parts  []string
		output string
	}{
		{[]string{"foo", "bar"}, "fooBar"},
		{[]string{"Foo", "bar"}, "fooBar"},
		{[]string{"Foo", "Bar"}, "fooBar"},
		{[]string{"fOo", "bAr"}, "fooBar"},

		{[]string{"foo", "url"}, "fooURL"},
		{[]string{"foo", "Url"}, "fooURL"},
		{[]string{"foo", "uRl"}, "fooURL"},
		{[]string{"foo", "URL"}, "fooURL"},
		{[]string{"url", "foo"}, "urlFoo"},

		{[]string{"foo", "cpu"}, "fooCPU"},
		{[]string{"foo", "Cpu"}, "fooCPU"},
		{[]string{"foo", "cPu"}, "fooCPU"},
		{[]string{"foo", "CPU"}, "fooCPU"},
		{[]string{"cpu", "foo"}, "cpuFoo"},

		{[]string{"foo", "ip"}, "fooIP"},
		{[]string{"foo", "Ip"}, "fooIP"},
		{[]string{"foo", "iP"}, "fooIP"},
		{[]string{"foo", "IP"}, "fooIP"},
		{[]string{"ip", "foo"}, "ipFoo"},

		{[]string{"foo", "id"}, "fooID"},
		{[]string{"foo", "Id"}, "fooID"},
		{[]string{"foo", "iD"}, "fooID"},
		{[]string{"foo", "ID"}, "fooID"},
		{[]string{"id", "foo"}, "idFoo"},
	}

	for _, sep := range []string{" ", ".", "_", "-", "--", "_-"} {
		for _, set := range sets {
			tests = append(tests, testStruct{strings.Join(set.parts, sep), set.output})
		}
	}

	for _, test := range tests {
		s.Equal(pflag.NormalizedName(test.output), configutil.NormalizeFunc(pflag.CommandLine, test.input), "'"+test.input+"'")
	}
}
