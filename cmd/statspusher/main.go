package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/spf13/pflag"
)

func main() {
	logrus.SetFormatter(&logrusx.MistifyFormatter{})

	config := newConfig(nil, nil)
	pflag.Parse()

	dieOnError(config.loadConfig())
	dieOnError(config.setupLogging())

	sp, err := newStatsPusher(config)
	dieOnError(err)

	dieOnError(sp.run())
	sp.stopOnSignal()
}

func dieOnError(err error) {
	if err != nil {
		logrus.Fatal("encountered an error during startup")
	}
}
