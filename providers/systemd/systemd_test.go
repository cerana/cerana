package systemd_test

import (
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
	svcPath, _ := filepath.Abs(filepath.Join("./_test", name))
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

func disable(name string) error {
	dconn, err := dbus.New()
	if err != nil {
		return err
	}
	defer dconn.Close()

	_, err = dconn.DisableUnitFiles([]string{name}, false)
	return err
}
