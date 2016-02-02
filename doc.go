/*
Package acomm is a library for asynchronous JSON-RPC-like communication between
services.

Like JSON-RPC, requests specify the task to run, arguments for the task, and an
id. Additionally, they specify a responseHook as well as success and error
handlers. The response hook is where the response should be sent, and the hooks
will run based on the type of response received. Responses are also similar to
JSON-RPC, with the addition of a StreamURL field, used to indicate that
additional data is available to stream directly. The request args and response
result are left as json.RawMessage and can be unmarshalled into whatever struct
the user desires.

The tracker provides request/response tracking. A request can be registered
with the tracker, along with a timeout, before sending. When a response
arrives, the request will be retrieved based on ID. Shutting down a tracker
will wait for all open requests to be handled, whether it is a response
arriving or a timeout occuring. The tracker also provides functionality for
proxying requests that use an http response hook to one that uses a unix socket
(provided by the tracker). It tracks the original request and returns a new
request using its response listener as the response hook. When the response
comes, it will then forward it along to the original response hook.

In a similar vein, the tracker can set up ad-hoc unix listeners for streaming
data, as well as proxy it to http. It includes an HTTP handler func for
handling http stream requests.

The UnixListener provides a wrapper around a unix socket, with connection
tracking for graceful shutdown. Communication over a unix socket is done by
sending a payload size header and then the JSON data; there are included
methods for handling the sending and reading of such data.
*/
package acomm
