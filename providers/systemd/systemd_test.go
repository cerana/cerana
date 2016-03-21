package systemd_test

import (
	"path/filepath"
	"testing"

	"github.com/coreos/go-systemd/dbus"
	"github.com/mistifyio/mistify/providers/systemd"
	"github.com/stretchr/testify/suite"
)

type sd struct {
	suite.Suite
	systemd *systemd.Systemd
}

func TestSd(t *testing.T) {
	suite.Run(t, new(sd))
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
