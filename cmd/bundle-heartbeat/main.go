package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/tick"
)

func main() {
	logrus.SetFormatter(&logrusx.JSONFormatter{})

	config := tick.NewConfig(nil, nil)
	logrusx.DieOnError(config.LoadConfig(), "load config")
	logrusx.DieOnError(config.SetupLogging(), "setup logging")

	stopChan, err := tick.RunTick(config, bundleHeartbeats)
	logrusx.DieOnError(err, "running tick")
	<-stopChan
}
