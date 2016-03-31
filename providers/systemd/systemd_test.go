package systemd_test

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cerana/cerana/provider"
	systemdp "github.com/cerana/cerana/providers/systemd"
	"github.com/coreos/go-systemd/dbus"
	"github.com/mistifyio/mistify-logrus-ext"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type systemd struct {
	suite.Suite
	dir     string
	config  *systemdp.Config
	systemd *systemdp.Systemd
	flagset *pflag.FlagSet
	viper   *viper.Viper
}

func TestSystemd(t *testing.T) {
	suite.Run(t, new(systemd))
}

func (s *systemd) SetupSuite() {
	dir, err := ioutil.TempDir("", "systemd-provider-test-")
	s.Require().NoError(err)
	s.dir = dir

	// Put premade unit files into dir provider was configured with so they can
	// be enabled by provider.
	files, err := ioutil.ReadDir("./_test_data")
	s.Require().NoError(err)
	for _, file := range files {
		oldPath, err := filepath.Abs(filepath.Join("./_test_data", file.Name()))
		s.Require().NoError(err)
		newPath, err := filepath.Abs(filepath.Join(s.dir, file.Name()))
		s.Require().NoError(err)
		s.Require().NoError(copyFile(oldPath, newPath))
	}

	v := viper.New()
	flagset := pflag.NewFlagSet("systemd", pflag.PanicOnError)
	config := systemdp.NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	v.Set("service_name", "systemd-provider-test")
	v.Set("socket_dir", s.dir)
	v.Set("coordinator_url", "unix:///tmp/foobar")
	v.Set("unit_file_dir", s.dir)
	v.Set("log_level", "fatal")
	s.Require().NoError(config.LoadConfig())
	s.Require().NoError(config.SetupLogging())
	s.config = config
	s.flagset = flagset
	s.viper = v

	s.systemd, err = systemdp.New(config)
	s.Require().NoError(err)
}

func (s *systemd) TearDownSuite() {
	_ = os.RemoveAll(s.dir)
}

func (s *systemd) TestRegisterTasks() {
	server, err := provider.NewServer(s.config.Config)
	s.Require().NoError(err)

	s.systemd.RegisterTasks(server)

	s.True(len(server.RegisteredTasks()) > 0)
}

func enable(name string) error {
	svcPath, err := filepath.Abs(filepath.Join("./_test_data", name))
	if err != nil {
		return err
	}
	dconn, err := dbus.New()
	if err != nil {
		return err
	}
	defer dconn.Close()

	if _, _, err = dconn.EnableUnitFiles([]string{svcPath}, false, true); err != nil {
		return err
	}

	return dconn.Reload()
}

func enableAll(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := enable(file.Name()); err != nil {
			return err
		}
	}
	return nil
}

func disable(name string) error {
	dconn, err := dbus.New()
	if err != nil {
		return err
	}
	defer dconn.Close()

	resultChan := make(chan string)
	_, err = dconn.StopUnit(name, systemdp.ModeFail, resultChan)
	result := <-resultChan
	if result != "done" {
		return errors.New(result)
	}

	_, err = dconn.DisableUnitFiles([]string{name}, false)
	return err
}

func disableAll(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		_ = disable(file.Name())
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer logrusx.LogReturnedErr(in.Close, nil, "failed to close source file")
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer logrusx.LogReturnedErr(out.Close, nil, "failed to close dest file")
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
