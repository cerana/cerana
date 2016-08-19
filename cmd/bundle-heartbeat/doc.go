/*
Usage:
	Usage of ./bundle-heartbeat:
	-u, --clusterDataURL string        url of coordinator for the cluster information
	-c, --configFile string            path to config file
	-l, --logLevel string              log level: debug/info/warn/error/fatal/panic (default "warning")
	-n, --nodeDataURL string           url of coordinator for node information retrieval
	-r, --requestTimeout duration      default timeout for external requests made
	-t, --tickInterval duration        tick run frequency
	-i, --tickRetryInterval duration   tick retry on error frequency
	Note: Long flag names can be specified in either fooBar or foo[_-.]bar form.
*/
package main
