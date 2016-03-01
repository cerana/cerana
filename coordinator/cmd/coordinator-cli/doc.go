/*
Usage

	$ ./coordinator-cli -h
	Usage of ./coordinator-cli:
	-c, --coordinator_url string   url of the coordinator
	-a, --request_arg value        task specific argument the form 'key:value'. can be set multiple times (default [])
	-r, --response_addr string     address for response http handler to listen on (default ":4080")
	-t, --task string              task to run
*/
package main
