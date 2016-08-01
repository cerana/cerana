# coordinator-cli

[![coordinator-cli](https://godoc.org/github.com/cerana/cerana/cmd/coordinator-cli?status.svg)](https://godoc.org/github.com/cerana/cerana/cmd/coordinator-cli)

### Usage

    $ ./coordinator-cli -h
    Usage of ./coordinator-cli:
    -c, --coordinator_url string   url of the coordinator
    -r, --http_addr string         address for http server to listen for responses and stream request data (default ":4080")
    -j, --json_args                read a json args object form STDIN
    -a, --request_arg value        task specific argument the form 'key=value'. can be set multiple times (default [])
    -s, --stream                   stream data from STDIN to provider
    -t, --task string              task to run
    -u, --task_url string          url of the task handler if different than coordinator


--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
