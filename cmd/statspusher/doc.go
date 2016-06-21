/*
Usage:
	$ statspusher -h
	Usage of ./statspusher:
	-b, --bundleInterval uint    bundle heartbeat interval (seconds)
	-c, --configFile string      path to config file
	-d, --datasetInterval uint   dataset heartbeat interval (seconds)
	-e, --heartbeatURL string    url of coordinator for the heartbeat registering
	-l, --logLevel string        log level: debug/info/warn/error/fatal/panic (default "warning")
	-u, --nodeDataURL string     url of coordinator for node information retrieval
	-n, --nodeInterval uint      node heartbeat interval (seconds)
	-r, --requestTimeout uint    default timeout for requests made (seconds)
	Note: Flags can be used in either fooBar or foo[_-.]bar form.
*/
package main
