/*
Package coordinator is a request router and proxy, to be used with Providers.
It is the entry point for all task requests and separates local providers from
external services.

New task requests are first sent to a Coordinator. There are two means the
Coordinator has of receiving requests: an http server, used for external
requests, and a unix socket, used for internal requests (i.e. from providers).
The Coordinator looks for providers of the task in the well defined directory
structure and sends the request to an appropriate provider. If the request has
an http response hook rather than a unix socket, it is tracked and a new
request using the Coordinator's response socket as the hook is sent to the
appropriate task. In this way, responses to external requests first come back
to the Coordinator and then go to the original response hook, rather than
Providers responding directly to the outside world. Similarly, if the response
to a proxied request contains a StreamURL, the Coordinator proxies the stream,
modifying the StreamURL being sent externally appropriately.

Endpoints

	External Request: http, /
	Internal Request: unix, /[socket_dir]/coordinator/[coordinator name].sock
	Internal Response: unix, /[socket_dir]/response/[coordinator name].sock
	Proxied Stream: http, /stream?addr=[original StreamURL]

Config

	{
		"config_file": "/path/to/config/file.json",
		"socket_dir": "/base/path/for/sockets",
		"service_name": "NameOfThisCoordinator",
		"external_port": 8080,
		"request_timeout": 0,
		"log_level": "warning"
	}
*/
package coordinator
