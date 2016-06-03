package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	logx "github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/kv"
	flag "github.com/spf13/pflag"
)

func main() {
	log.SetFormatter(&logx.JSONFormatter{})

	config := kv.NewConfig(nil, nil)
	flag.StringP("address", "a", "", "kv address (leave blank for default)")
	flag.Parse()

	dieOnError(config.LoadConfig())
	dieOnError(config.SetupLogging())

	server, err := provider.NewServer(config.Config)
	dieOnError(err)

	k, err := kv.New(config, server.Tracker())
	dieOnError(err)
	k.RegisterTasks(server)

	if len(server.RegisteredTasks()) == 0 {
		log.Warn("no registered tasks, exiting")
		os.Exit(1)
	}
	dieOnError(server.Start())
	server.StopOnSignal()
}

func dieOnError(err error) {
	if err != nil {
		log.Fatal("encountered an error during startup, error:", err)
		os.Exit(1)
	}
}
