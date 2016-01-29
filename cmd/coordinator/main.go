package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/coordinator"
	logx "github.com/mistifyio/mistify-logrus-ext"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	log.SetFormatter(&logx.MistifyFormatter{})

	flag.StringP("config_file", "c", "", "path to config file")
	flag.StringP("service_name", "n", "", "name of the coordinator")
	flag.StringP("socket_dir", "s", "/tmp/mistify", "base directory in which to create task sockets")
	flag.IntP("external_port", "p", 8080, "port for the http external request server to listen")
	flag.StringP("log_level", "l", "warning", "log level: debug/info/warn/error/fatal/panic")
	flag.Uint64P("request_timeout", "t", 0, "default timeout for requests in seconds")
	flag.Parse()

	v := viper.New()
	bindFlags(v)

	v.SetDefault("service_name", "coordinator")

	config := coordinator.NewConfig(v)
	dieOnError(config.LoadConfigFile())
	dieOnError(config.SetupLogging())

	server, err := coordinator.NewServer(config)
	dieOnError(err)

	dieOnError(server.Start())
	server.StopOnSignal()
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
