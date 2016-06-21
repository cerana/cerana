/*
Usage:
	$ statspusher -h
	Usage of statspusher:
	-b, --bundleTTL uint        bundle heartbeat ttl (seconds)
	-c, --configFile string     path to config file
	-d, --datasetTTL uint       dataset heartbeat ttl (seconds)
	-e, --heartbeatURL string   url of coordinator for the heartbeat registering
	-l, --logLevel string       log level: debug/info/warn/error/fatal/panic (default "warning")
	-u, --nodeDataURL string    url of coordinator for node information retrieval
	-n, --nodeTTL uint          node heartbeat ttl (seconds)
	-r, --requestTimeout uint   default timeout for requests made (seconds)
	Note: Flags can be used in either fooBar or foo[_-.]bar form.
*/
package main
