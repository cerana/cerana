package systemd_test

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/coreos/go-systemd/dbus"
	systemdp "github.com/mistifyio/mistify/providers/systemd"
	"github.com/stretchr/testify/suite"
)

type systemd struct {
	suite.Suite
	systemd *systemdp.Systemd
}

func TestSystemd(t *testing.T) {
	suite.Run(t, new(systemd))
}

func enable(name string) error {
	svcPath, err := filepath.Abs(filepath.Join("./_test", name))
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
