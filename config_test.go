package provider_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/provider"
	"github.com/pborman/uuid"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type ConfigSuite struct {
	suite.Suite
}

func (s *ConfigSuite) SetupTest() {
	log.SetLevel(log.FatalLevel)
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

func (s *ConfigSuite) TestNewConfig() {
	tests := []struct {
		description string
		flagSet     *flag.FlagSet
		viper       *viper.Viper
	}{
		{"defaults", nil, nil},
		{"specified flagset", flag.NewFlagSet(uuid.New(), flag.ExitOnError), nil},
		{"specified viper", nil, viper.New()},
		{"specified both", flag.NewFlagSet(uuid.New(), flag.ExitOnError), viper.New()},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		// Reset the default CommandLine flags between runs
		flag.CommandLine = flag.NewFlagSet(uuid.New(), flag.ExitOnError)
		s.NotNil(provider.NewConfig(test.flagSet, test.viper), msg("failed to return a config"))
	}
}

func (s *ConfigSuite) TestLoadConfig() {
	data := struct {
		SocketDir      string `json:"socket_dir"`
		ServiceName    string `json:"service_name"`
		CoordinatorURL string `json:"coordinator_url"`
	}{
		os.TempDir(),
		uuid.New(),
		"http://localhost.com:8080/",
	}

	dataJSON, _ := json.Marshal(data)
	fmt.Println(string(dataJSON))

	tests := []struct {
		description string
		flagSet     bool
		configFile  bool
		expectedErr bool
	}{
		{"nothing set", false, false, true},
		{"flags only", true, false, false},
		{"config file", false, true, false},
		{"flags and config", true, true, false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)

		fs := flag.NewFlagSet(uuid.New(), flag.ExitOnError)
		v := viper.New()
		v.SetConfigType("json")
		config := provider.NewConfig(fs, v)
		if !s.NotNil(config, msg("failed to return a config")) {
			continue
		}

		configFileName := ""
		if test.configFile {
			configFile, err := ioutil.TempFile("", "providerTest-")
			if !s.NoError(err, msg("failed to create conf file")) {
				continue
			}
			configFileName = configFile.Name()
			defer os.Remove(configFileName)
			defer configFile.Close()

			if _, err := configFile.Write(dataJSON); !s.NoError(err, msg("failed to write config data")) {
				continue
			}

			fs.Set("config_file", configFileName)
		}

		if test.flagSet {
			fs.Set("socket_dir", data.SocketDir)
			fs.Set("service_name", data.ServiceName)
			fs.Set("coordinator_url", data.CoordinatorURL)
		}

		sn, err := fs.GetString("service_name")
		fmt.Println(msg("should be '%s', sn '%s', err %s", data.ServiceName, sn, err))
		/*
			if !s.NoError(fs.Parse(), msg("failed to parse flags")) {
				continue
			}
		*/

		if test.expectedErr {
			s.Error(config.LoadConfig(), msg("should not load valid config"))
		} else {
			s.NoError(config.LoadConfig(), msg("failed to load valid config"))
		}
	}
}

func (s *ConfigSuite) TestTaskPriority() {
}

func (s *ConfigSuite) TestTaskTimeout() {
}

func (s *ConfigSuite) TestSocketDir() {
}

func (s *ConfigSuite) TestStreamDir() {
}

func (s *ConfigSuite) TestServiceName() {
}

func (s *ConfigSuite) TestCoordinatorURL() {
}

func (s *ConfigSuite) TestRequestTimeout() {
}

func (s *ConfigSuite) TestValidate() {
}

func (s *ConfigSuite) TestUnmarshal() {
}

func (s *ConfigSuite) TestUnmarshalKey() {
}

func (s *ConfigSuite) TestSetupLogging() {
}

func testMsgFunc(prefix string) func(...interface{}) string {
	return func(val ...interface{}) string {
		if len(val) == 0 {
			return prefix
		}
		msgPrefix := prefix + " : "
		if len(val) == 1 {
			return msgPrefix + val[0].(string)
		} else {
			return msgPrefix + fmt.Sprintf(val[0].(string), val[1:]...)
		}
	}
}
