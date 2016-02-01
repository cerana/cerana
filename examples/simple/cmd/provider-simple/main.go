package main

import (
	log "github.com/Sirupsen/logrus"
	logx "github.com/mistifyio/mistify-logrus-ext"
	"github.com/mistifyio/provider"
	"github.com/mistifyio/provider/examples/simple"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	log.SetFormatter(&logx.MistifyFormatter{})

	flag.StringP("config_file", "c", "", "path to config file")
	flag.StringP("socket_dir", "s", "/tmp/mistify", "base directory in which to create task sockets")
	flag.UintP("default_priority", "p", 50, "default task priority")
	flag.StringP("log_level", "l", "warning", "log level: debug/info/warn/error/fatal/panic")
	flag.Uint64P("request_timeout", "t", 0, "default timeout for requests made by this provider in seconds")
	flag.Parse()

	v := viper.New()
	bindFlags(v)

	v.SetDefault("service_name", "provider-simple")

	config := provider.NewConfig(v)
	dieOnError(config.LoadConfigFile())
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

// Bind the command line flags to Viper so they will be merged into the config
func bindFlags(v *viper.Viper) {
	if err := v.BindPFlags(flag.CommandLine); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("failed to bind pflags to viper")
	}
}

func dieOnError(err error) {
	if err != nil {
		log.Fatal("encountered an error during startup")
	}
}
