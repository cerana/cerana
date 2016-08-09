package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/spf13/pflag"
)

func main() {
	logrus.SetFormatter(&logrusx.JSONFormatter{})

	config := newConfig(nil, nil)
	pflag.Parse()

	logrusx.DieOnError(config.loadConfig(), "load config")
	logrusx.DieOnError(config.setupLogging(), "setup logging")

	sp, err := newStatsPusher(config)
	logrusx.DieOnError(err, "new statspusher")

	logrusx.DieOnError(sp.run(), "run statspusher")
	sp.stopOnSignal()
}
