/*
Usage

	$ service-provider -h
	Usage of service-provider:
	-c, --config_file string       path to config file
	-u, --coordinator_url string   url of coordinator for making requests
	-p, --default_priority uint    default task priority (default 50)
	-l, --log_level string         log level: debug/info/warn/error/fatal/panic (default "warning")
	-t, --request_timeout uint     default timeout for requests made by this provider in seconds
	-n, --service_name string      provider service name
	-s, --socket_dir string        base directory in which to create task sockets (default "/tmp/cerana")
*/
package main
