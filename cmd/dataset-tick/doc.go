/*
Usage
	$ dataset-tick -h
	Usage of dataset-tick:
	-u, --clusterDataURL string        url of coordinator for the cluster information
	-c, --configFile string            path to config file
	-d, --datasetPrefix string         dataset directory
	-l, --logLevel string              log level: debug/info/warn/error/fatal/panic (default "warning")
	-n, --nodeDataURL string           url of coordinator for node information retrieval
	-r, --requestTimeout duration      default timeout for external requests made
	-a, --responseAddr string          address on which to listen for http responses
	-t, --tickInterval duration        tick run frequency
	-i, --tickRetryInterval duration   tick retry on error frequency
	Note: Long flag names can be specified in either fooBar or foo[_-.]bar form.
*/
package main
