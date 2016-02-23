package provider_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/provider"
	"github.com/pborman/uuid"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

func TestConfig(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

type ConfigSuite struct {
	suite.Suite
	config     *provider.Config
	configData *provider.ConfigData
	configFile *os.File
}

func (s *ConfigSuite) SetupTest() {
	log.SetLevel(log.FatalLevel)

	s.configData = &provider.ConfigData{
		SocketDir:       os.TempDir(),
		ServiceName:     uuid.New(),
		CoordinatorURL:  "http://localhost:8080/",
		DefaultPriority: 43,
		LogLevel:        "fatal",
		DefaultTimeout:  100,
		RequestTimeout:  10,
		Tasks: map[string]*provider.TaskConfigData{
			"foobar": &provider.TaskConfigData{
				Priority: 56,
				Timeout:  64,
			},
		},
	}

	var err error
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
		s.NotNil(provider.NewConfig(test.flagSet, test.viper), msg("failed to return a config"))
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

func (s *ConfigSuite) TestTaskPriority() {
	s.EqualValues(s.configData.DefaultPriority, s.config.TaskPriority(uuid.New()))
	s.EqualValues(s.configData.Tasks["foobar"].Priority, s.config.TaskPriority("foobar"))
}

func (s *ConfigSuite) TestTaskTimeout() {
	s.EqualValues(s.configData.DefaultTimeout, s.config.TaskTimeout(uuid.New())/time.Second)
	s.EqualValues(s.configData.Tasks["foobar"].Timeout, s.config.TaskTimeout("foobar")/time.Second)
}

func (s *ConfigSuite) TestSocketDir() {
	s.Equal(s.configData.SocketDir, s.config.SocketDir())
}

func (s *ConfigSuite) TestStreamDir() {
	dir := s.config.StreamDir("foobar")
	s.Contains(dir, s.config.SocketDir(), "missing socket dir")
	s.Contains(dir, "foobar", "missing task name")
	s.Contains(dir, s.config.ServiceName(), "missing service name")
}

func (s *ConfigSuite) TestServiceName() {
	s.Equal(s.configData.ServiceName, s.config.ServiceName())
}

func (s *ConfigSuite) TestCoordinatorURL() {
	s.Equal(s.configData.CoordinatorURL, s.config.CoordinatorURL().String())
}

func (s *ConfigSuite) TestRequestTimeout() {
	s.EqualValues(s.configData.RequestTimeout, s.config.RequestTimeout()/time.Second)
}

func (s *ConfigSuite) TestValidate() {
	tests := []struct {
		description    string
		socketDir      string
		serviceName    string
		coordinatorURL string
		expectedError  bool
	}{
		{"valid", "/tmp", "foobar", "http://localhost:8080/", false},
		{"missing socket dir", "", "foobar", "http://localhost:8080/", true},
		{"missing service name", "/tmp", "", "http://localhost:8080/", true},
		{"missing coordinator url", "/tmp", "foobar", "", true},
		{"bad coordinator url", "/tmp", "foobar", "=aer=0./", true},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)

		configData := &provider.ConfigData{
			SocketDir:      test.socketDir,
			ServiceName:    test.serviceName,
			CoordinatorURL: test.coordinatorURL,
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

func (s *ConfigSuite) TestUnmarshal() {
	config := &provider.ConfigData{}
	if !s.NoError(s.config.Unmarshal(config)) {
		return
	}
	s.Equal(s.configData, config)
}

func (s *ConfigSuite) TestUnmarshalKey() {
	taskConfig := &provider.TaskConfigData{}
	if !s.NoError(s.config.UnmarshalKey("tasks.foobar", taskConfig)) {
		return
	}
	s.Equal(s.configData.Tasks["foobar"], taskConfig)
}

func (s *ConfigSuite) TestSetupLogging() {
	s.NoError(s.config.SetupLogging(), "failed to setup logging")
	s.Equal(s.configData.LogLevel, log.GetLevel().String())
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

func newConfig(setFlags, writeConfig bool, configData *provider.ConfigData) (*provider.Config, *flag.FlagSet, *viper.Viper, *os.File, error) {
	fs := flag.NewFlagSet(uuid.New(), flag.ExitOnError)
	v := viper.New()
	v.SetConfigType("json")
	config := provider.NewConfig(fs, v)
	if config == nil {
		return nil, nil, nil, nil, errors.New("failed to return a config")
	}

	var configFile *os.File
	if writeConfig {
		var err error
		configFile, err = ioutil.TempFile("", "providerTest-")
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
		if err := fs.Set("coordinator_url", configData.CoordinatorURL); err != nil {
			return nil, nil, nil, configFile, err
		}
	}

	return config, fs, v, configFile, nil
}
