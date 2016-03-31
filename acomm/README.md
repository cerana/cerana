# acomm

[![acomm](https://godoc.org/github.com/cerana/cerana/acomm?status.png)](https://godoc.org/github.com/cerana/cerana/acomm)

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

The tracker provides request/response tracking. A request can be registered with
the tracker, along with a timeout, before sending. When a response arrives, the
request will be retrieved based on ID. Shutting down a tracker will wait for all
open requests to be handled, whether it is a response arriving or a timeout
occuring. The tracker also provides functionality for proxying requests that use
an http response hook to one that uses a unix socket (provided by the tracker).
It tracks the original request and returns a new request using its response
listener as the response hook. When the response comes, it will then forward it
along to the original response hook.

In a similar vein, the tracker can set up ad-hoc unix listeners for streaming
data, as well as proxy it to http. It includes an HTTP handler func for handling
http stream requests.

The UnixListener provides a wrapper around a unix socket, with connection
tracking for graceful shutdown. Communication over a unix socket is done by
sending a payload size header and then the JSON data; there are included methods
for handling the sending and reading of such data.

## Usage

#### func  ProxyStreamHandler

```go
func ProxyStreamHandler(w http.ResponseWriter, r *http.Request)
```
ProxyStreamHandler is an HTTP HandlerFunc for simple proxy streaming.

#### func  Send

```go
func Send(addr *url.URL, payload interface{}) error
```
Send attempts send the payload to the specified URL.

#### func  SendConnData

```go
func SendConnData(conn net.Conn, payload interface{}) error
```
SendConnData marshals and writes payload JSON data to the Conn with appropriate
headers.

#### func  Stream

```go
func Stream(dest io.Writer, addr *url.URL) error
```
Stream streams data from a URL to a destination writer.

#### func  UnmarshalConnData

```go
func UnmarshalConnData(conn net.Conn, dest interface{}) error
```
UnmarshalConnData reads and unmarshals JSON data from the connection into the
destination object.

#### type MultiRequest

```go
type MultiRequest struct {
}
```

MultiRequest provides a way to manage multiple parallel requests

#### func  NewMultiRequest

```go
func NewMultiRequest(tracker *Tracker, timeout time.Duration) *MultiRequest
```
NewMultiRequest creates and initializes a new MultiRequest.

#### func (*MultiRequest) AddRequest

```go
func (m *MultiRequest) AddRequest(name string, req *Request) error
```
AddRequest adds a request to the MultiRequest. Sending the request is still the
responsibility of the caller.

#### func (*MultiRequest) RemoveRequest

```go
func (m *MultiRequest) RemoveRequest(req *Request)
```
RemoveRequest removes a request from the MultiRequest. Useful if the send fails.

#### func (*MultiRequest) Responses

```go
func (m *MultiRequest) Responses() map[string]*Response
```
Responses returns responses for all of the requests, keyed on the request name
(as opposed to request id). Blocks until all requests are accounted for.

#### type Request

```go
type Request struct {
	ID             string           `json:"id"`
	Task           string           `json:"task"`
	ResponseHook   *url.URL         `json:"responsehook"`
	StreamURL      *url.URL         `json:"stream_url"`
	Args           *json.RawMessage `json:"args"`
	SuccessHandler ResponseHandler  `json:"-"`
	ErrorHandler   ResponseHandler  `json:"-"`
}
```

Request is a request data structure for asynchronous requests. The ID is used to
identify the request throught its life cycle. The ResponseHook is a URL where
response data should be sent. SuccessHandler and ErrorHandler will be called
appropriately to handle a response.

#### func  NewRequest

```go
func NewRequest(task, responseHook, streamURL string, args interface{}, sh ResponseHandler, eh ResponseHandler) (*Request, error)
```
NewRequest creates a new Request instance.

#### func (*Request) HandleResponse

```go
func (req *Request) HandleResponse(resp *Response)
```
HandleResponse determines whether a response indicates success or error and runs
the appropriate handler. If the appropriate handler is not defined, it is
assumed no handling is necessary and silently finishes.

#### func (*Request) Respond

```go
func (req *Request) Respond(resp *Response) error
```
Respond sends a Response to a Request's ResponseHook.

#### func (*Request) UnmarshalArgs

```go
func (req *Request) UnmarshalArgs(dest interface{}) error
```
UnmarshalArgs unmarshals the request args into the destination object.

#### func (*Request) Validate

```go
func (req *Request) Validate() error
```
Validate validates the reqeust

#### type Response

```go
type Response struct {
	ID        string           `json:"id"`
	Result    *json.RawMessage `json:"result"`
	StreamURL *url.URL         `json:"stream_url"`
	Error     error            `json:"error"`
}
```

Response is a response data structure for asynchronous requests. The ID should
be the same as the Request it corresponds to. Result should be nil if Error is
present and vice versa.

#### func  NewResponse

```go
func NewResponse(req *Request, result interface{}, streamURL *url.URL, respErr error) (*Response, error)
```
NewResponse creates a new Response instance based on a Request.

#### func (*Response) MarshalJSON

```go
func (r *Response) MarshalJSON() ([]byte, error)
```
MarshalJSON marshals a Response into JSON.

#### func (*Response) UnmarshalJSON

```go
func (r *Response) UnmarshalJSON(data []byte) error
```
UnmarshalJSON unmarshals JSON data into a Response.

#### func (*Response) UnmarshalResult

```go
func (r *Response) UnmarshalResult(dest interface{}) error
```
UnmarshalResult unmarshals the response result into the destination object.

#### type ResponseHandler

```go
type ResponseHandler func(*Request, *Response)
```

ResponseHandler is a function to run when a request receives a response.

#### type Tracker

```go
type Tracker struct {
}
```

Tracker keeps track of requests waiting on a response.

#### func  NewTracker

```go
func NewTracker(socketPath string, httpStreamURL *url.URL, defaultTimeout time.Duration) (*Tracker, error)
```
NewTracker creates and initializes a new Tracker. If a socketPath is not
provided, the response socket will be created in a temporary directory.

#### func (*Tracker) Addr

```go
func (t *Tracker) Addr() string
```
Addr returns the string representation of the Tracker's response listener
socket.

#### func (*Tracker) HandleResponse

```go
func (t *Tracker) HandleResponse(resp *Response)
```
HandleResponse associates a response with a request and either forwards the
response or calls the request's handler.

#### func (*Tracker) NewStreamUnix

```go
func (t *Tracker) NewStreamUnix(dir string, src io.ReadCloser) (*url.URL, error)
```
NewStreamUnix sets up an ad-hoc unix listner to stream data.

#### func (*Tracker) NumRequests

```go
func (t *Tracker) NumRequests() int
```
NumRequests returns the number of tracked requests

#### func (*Tracker) ProxyStreamHTTPURL

```go
func (t *Tracker) ProxyStreamHTTPURL(addr *url.URL) (*url.URL, error)
```
ProxyStreamHTTPURL generates the url for proxying streaming data from a unix
socket.

#### func (*Tracker) ProxyUnix

```go
func (t *Tracker) ProxyUnix(req *Request, timeout time.Duration) (*Request, error)
```
ProxyUnix proxies requests that have response hooks and stream urls of non-unix
sockets. If the response hook and stream url are already unix sockets, it
returns the original request. If the response hook is not, it tracks the
original request and returns a new request with a unix socket response hook. If
the stream url is not, it pipes the original stream through a new unix socket
and updates the stream url. The purpose of this is so that there can be a single
entry and exit point for external communication, while local services can reply
directly to each other.

#### func (*Tracker) RemoveRequest

```go
func (t *Tracker) RemoveRequest(req *Request) bool
```
RemoveRequest should be used to remove a tracked request. Use in cases such as
sending failures, where there is no hope of a response being received.

#### func (*Tracker) Start

```go
func (t *Tracker) Start() error
```
Start activates the tracker. This allows tracking of requests as well as
listening for and handling responses.

#### func (*Tracker) Stop

```go
func (t *Tracker) Stop()
```
Stop deactivates the tracker. It blocks until all active connections or tracked
requests to finish.

#### func (*Tracker) TrackRequest

```go
func (t *Tracker) TrackRequest(req *Request, timeout time.Duration) error
```
TrackRequest tracks a request. This does not need to be called after using
ProxyUnix.

#### func (*Tracker) URL

```go
func (t *Tracker) URL() *url.URL
```
URL returns the URL of the Tracker's response listener socket.

#### type UnixListener

```go
type UnixListener struct {
}
```

UnixListener is a wrapper for a unix socket. It handles creation and listening
for new connections, as well as graceful shutdown.

#### func  NewUnixListener

```go
func NewUnixListener(socketPath string, acceptLimit int) *UnixListener
```
NewUnixListener creates and initializes a new UnixListener. AcceptLimit controls
how many connections it will listen for before stopping; 0 and below is
unlimited.

#### func (*UnixListener) Addr

```go
func (ul *UnixListener) Addr() string
```
Addr returns the string representation of the unix address.

#### func (*UnixListener) DoneConn

```go
func (ul *UnixListener) DoneConn(conn net.Conn)
```
DoneConn completes the handling of a connection.

#### func (*UnixListener) NextConn

```go
func (ul *UnixListener) NextConn() net.Conn
```
NextConn blocks and returns the next connection. It will return nil when the
listener is stopped and all existing connections have been handled. Connections
should be handled in a go routine to take advantage of concurrency. When done,
the connection MUST be finished with a call to DoneConn.

#### func (*UnixListener) Start

```go
func (ul *UnixListener) Start() error
```
Start prepares the listener and starts listening for new connections.

#### func (*UnixListener) Stop

```go
func (ul *UnixListener) Stop(timeout time.Duration)
```
Stop stops listening for new connections. It blocks until existing connections
are handled and the listener closed.

#### func (*UnixListener) URL

```go
func (ul *UnixListener) URL() *url.URL
```
URL returns the URL representation of the unix address.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
