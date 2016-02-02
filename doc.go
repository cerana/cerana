/*
Package provider is the framework for creating Providers. A Provider offers a
set of related functionality in the form of tasks, which Coordinators use
individually or in combination, to accomplish actions.

Functionality Registration

Coordinators need to know what providers are available and what they offer in
order to complete tasks. Instead of an active registration, which requires
heartbeats and deregistration, Providers utilize the filesystem.  For
registration of each task, a unix socket is created in a known directory
structure based on the task name. A Coordinator can check a task directory for
the socket when routing requests. When a Provider is shut down, the unix socket
will be automatically removed, acting as the de-registration in the eyes of the
Coordinator. In order to handle multiple Providers capable of handling the same
task, socket filenames are prefixed with a priority value.

Task socket path: /[socket_dir]/[task_name]/[priority]-[provider-name].sock

Communication

Communication is handled via the acomm package. An initial request is received
on a task-specific socket. A response is sent to acknowledge receipt on the
same connection before the request is sent to the appropriate TaskHandler.
After completing the work, the TaskHandler returns a combination of result,
data stream url, and error. These three are bundled into a response and sent to
the request's responseHook. In the case of data streaming, the caller will
connect to the stream url and stream the data.

All requests originating from a provider go through the coordinator; providers
should not make requests directly to each other. These requests should be
tracked with the tracker. Responses may be sent directly to a unix socket
response hook. Data streaming is also handled directly. While providers could
respond to external (http) response hooks and stream data directly externally,
they will be separated by the coordinator, through which responses will be
proxied. Providers are free, however, to make requests to external services.

Creating A Provider

As the description states, the provider package is not itself a complete
provider; it is used to create one. The process is quite simple. A provider
implementation will have a set of functions following the TaskHandler
signature, which accepts a request and returns components for a response.  If
any tasks are going to make additional requests, the provider server's tracker
should be used. The provider must follow the Provider interface, which requires
a method to register the task handlers; the method should accept a Server and
call its RegisterTask method for each task handler.

Create a new Config, optionally supplying a flagset or viper instance. If
flagset is not provided, it will use the commandline flagset, and if viper is
not provided, one will be created. Flags should then be parsed, and the Config
loaded. After loading the Config, create a new Server using it. Then
create/initialize the provider, and register the tasks. Finally, start the
server and either use StopOnSignal or another way to wait and Stop the server
appropriately.

An example provider can be found in the examples directory.

Config

There are a number of values required in the config for a provider to operate successfully. The Config struct will add a number of the config options as flags (including `config_file`). Tasks without explicit config for priority will use default value.

	{
		"config_file": "/path/to/config/file.json",
		"socket_dir": "/base/path/for/sockets",
		"default_priority": 50,
		"log_level": "warning",
		"request_timeout": 0,
		"tasks":{
			"ATaskNameFoo":{
				"priority": 60,
			}
		}
	}

Suggestions

Task handlers should be kept focused and self-contained as possible, doing one
logical operation. Use additional task requests to compose these focused
building blocks into actions.

All non-informational tasks should have a corrosponding reverse or cleanup
task. Make use of request error handlers to call such cleanup tasks.
*/
package provider
