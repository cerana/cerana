package coordinator_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/coordinator"
	"github.com/pborman/uuid"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

type ConfigSuite struct {
	suite.Suite
	config     *coordinator.Config
	configData *coordinator.ConfigData
	configFile *os.File
}

func (s *ConfigSuite) SetupTest() {
	logrus.SetLevel(logrus.FatalLevel)

	socketDir, err := ioutil.TempDir("", "coordinatorTest-")
	s.Require().NoError(err, "failed to create socket dir")

	s.configData = &coordinator.ConfigData{
		SocketDir:      socketDir,
		ServiceName:    uuid.New(),
		ExternalPort:   45678,
		RequestTimeout: 5,
		LogLevel:       "fatal",
	}

	s.config, _, _, s.configFile, err = newConfig(false, true, s.configData)
	s.Require().NoError(err, "failed to create config")
	s.Require().NoError(s.config.LoadConfig(), "failed to load config")
}

func (s *ConfigSuite) TearDownTest() {
	_ = os.Remove(s.configFile.Name())
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
		s.NotNil(coordinator.NewConfig(test.flagSet, test.viper), msg("failed to return a config"))
	}
}

func (s *ConfigSuite) TestLoadConfig() {
	tests := []struct {
		description string
		setFlags    bool
		writeConfig bool
		expectedErr bool
	}{
		{"nothing set", false, false, true},
		{"flags only", true, false, false},
		{"config file", false, true, false},
		{"flags and config", true, true, false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)

		config, _, _, configFile, err := newConfig(test.setFlags, test.writeConfig, s.configData)
		if configFile != nil {
			defer func() { _ = os.Remove(configFile.Name()) }()
		}
		if !s.NoError(err, msg("failed to create and load config")) {
			continue
		}

		if test.expectedErr {
			s.Error(config.LoadConfig(), msg("should not load valid config"))
		} else {
			s.NoError(config.LoadConfig(), msg("failed to load valid config"))
		}
	}
}

func (s *ConfigSuite) TestSocketDir() {
	s.Equal(s.configData.SocketDir, s.config.SocketDir())
}

func (s *ConfigSuite) TestServiceName() {
	s.Equal(s.configData.ServiceName, s.config.ServiceName())
}

func (s *ConfigSuite) ExternalPort() {
	s.Equal(s.configData.ExternalPort, s.config.ExternalPort())
}

func (s *ConfigSuite) TestRequestTimeout() {
	s.EqualValues(s.configData.RequestTimeout, s.config.RequestTimeout()/time.Second)
}

func (s *ConfigSuite) TestValidate() {
	tests := []struct {
		description   string
		socketDir     string
		serviceName   string
		externalPort  uint
		expectedError bool
	}{
		{"valid", "/tmp", "foobar", 8080, false},
		{"missing socket dir", "", "foobar", 8080, true},
		{"missing service name", "/tmp", "", 8080, true},
		{"missing external port", "/tmp", "foobar", 0, true},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)

		configData := &coordinator.ConfigData{
			SocketDir:    test.socketDir,
			ServiceName:  test.serviceName,
			ExternalPort: test.externalPort,
		}

		config, fs, v, _, err := newConfig(true, false, configData)
		if !s.NoError(err, msg("failed to create config")) {
			continue
		}
		// Bind here to avoid the need for Load
		_ = v.BindPFlags(fs)

		if test.expectedError {
			s.Error(config.Validate(), msg("should not be valid"))
		} else {
			s.NoError(config.Validate(), msg("should be valid"))
		}
	}
}

func (s *ConfigSuite) TestSetupLogging() {
	s.NoError(s.config.SetupLogging(), "failed to setup logging")
	s.Equal(s.configData.LogLevel, logrus.GetLevel().String())
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

func newConfig(setFlags, writeConfig bool, configData *coordinator.ConfigData) (*coordinator.Config, *flag.FlagSet, *viper.Viper, *os.File, error) {
	fs := flag.NewFlagSet(uuid.New(), flag.ExitOnError)
	v := viper.New()
	v.SetConfigType("json")
	config := coordinator.NewConfig(fs, v)
	if config == nil {
		return nil, nil, nil, nil, errors.New("failed to return a config")
	}

	var configFile *os.File
	if writeConfig {
		var err error
		configFile, err = ioutil.TempFile("", "coordinatorTest-")
		if err != nil {
			return nil, nil, nil, nil, err
		}
		defer func() { _ = configFile.Close() }()

		configJSON, _ := json.Marshal(configData)
		if _, err := configFile.Write(configJSON); err != nil {
			return nil, nil, nil, configFile, err
		}

		if err := fs.Set("config_file", configFile.Name()); err != nil {
			return nil, nil, nil, configFile, err
		}
	}

	if setFlags {
		if err := fs.Set("socket_dir", configData.SocketDir); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("service_name", configData.ServiceName); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("external_port", strconv.FormatUint(uint64(configData.ExternalPort), 10)); err != nil {
			return nil, nil, nil, configFile, err
		}
	}

	return config, fs, v, configFile, nil
}
