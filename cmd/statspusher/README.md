# statspusher

[![statspusher](https://godoc.org/github.com/cerana/cerana/cmd/statspusher?status.svg)](https://godoc.org/github.com/cerana/cerana/cmd/statspusher)

Usage:

    $ ./statspusher -h
    Usage of ./statspusher:
    -b, --bundleInterval uint     bundle heartbeat interval
    -e, --clusterDataURL string   url of coordinator for the cluster information
    -c, --configFile string       path to config file
    -d, --datasetInterval uint    dataset heartbeat interval
    -l, --logLevel string         log level: debug/info/warn/error/fatal/panic (default "warning")
    -u, --nodeDataURL string      url of coordinator for node information retrieval
    -n, --nodeInterval uint       node heartbeat interval
    -r, --requestTimeout uint     default timeout for requests made
    Note: Flags can be used in either fooBar or foo[_-.]bar form.


--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
