# coordinator

[![coordinator](https://godoc.org/github.com/mistifyio/coordinator/cmd/coordinator?status.png)](https://godoc.org/github.com/mistifyio/coordinator/cmd/coordinator)

### Usage

    $ coordinator -h
    Usage of coordinator:
    -c, --config_file="": path to config file
    -p, --external_port=8080: port for the http external request server to listen
    -l, --log_level="warning": log level: debug/info/warn/error/fatal/panic
    -t, --request_timeout=0: default timeout for requests in seconds
    -n, --service_name="": name of the coordinator
    -s, --socket_dir="/tmp/mistify": base directory in which to create task sockets


--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
