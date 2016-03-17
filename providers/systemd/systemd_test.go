package systemd_test

import (
	"testing"

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
