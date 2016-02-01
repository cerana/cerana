package main

import (
	log "github.com/Sirupsen/logrus"
	logx "github.com/mistifyio/mistify-logrus-ext"
	"github.com/mistifyio/provider"
	"github.com/mistifyio/provider/examples/simple"
	flag "github.com/spf13/pflag"
)

func main() {
	log.SetFormatter(&logx.MistifyFormatter{})

	config := provider.NewConfig(nil, nil)

	flag.StringP("foobar", "f", "", "path to config file")

	flag.Parse()

	dieOnError(config.LoadConfig())
	dieOnError(config.SetupLogging())

	server, err := provider.NewServer(config)
	dieOnError(err)
	s := simple.NewSimple(config, server.Tracker())
	s.RegisterTasks(server)

	if len(server.RegisteredTasks()) != 0 {
		dieOnError(server.Start())
		server.StopOnSignal()
	} else {
		log.Warn("no registered tasks, exiting")
	}
}

func dieOnError(err error) {
	if err != nil {
		log.Fatal("encountered an error during startup")
	}
}
