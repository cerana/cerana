# acomm

[![acomm](https://godoc.org/github.com/mistifyio/acomm?status.png)](https://godoc.org/github.com/mistifyio/acomm)

Package acomm is a library for asynchronous communication between services.

## Usage

#### func  Send

```go
func Send(addr *url.URL, payload interface{}) error
```
Send attempts send the payload to the specified URL.

#### func  UnmarshalConnData

```go
func UnmarshalConnData(conn net.Conn, dest interface{}) error
```
UnmarshalConnData reads and unmarshals JSON data from the connection into the
destination object.

#### type Request

```go
type Request struct {
	ID             string          `json:"id"`
	Task           string          `json:"task"`
	ResponseHook   *url.URL        `json:"responsehook"`
	Args           interface{}     `json:"args"`
	SuccessHandler ResponseHandler `json:"-"`
	ErrorHandler   ResponseHandler `json:"-"`
}
```

Request is a request data structure for asynchronous requests. The ID is used to
identify the request throught its life cycle. The ResponseHook is a URL where
response data should be sent. SuccessHandler and ErrorHandler will be called
appropriately to handle a response.

#### func  NewRequest

```go
func NewRequest(task, responseHook string, args interface{}, sh ResponseHandler, eh ResponseHandler) (*Request, error)
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

#### func (*Request) Validate

```go
func (req *Request) Validate() error
```
Validate validates the reqeust

#### type Response

```go
type Response struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result"`
	Error  error       `json:"error"`
}
```

Response is a response data structure for asynchronous requests. The ID should
be the same as the Request it corresponds to. Result should be nil if Error is
present and vice versa.

#### func  NewResponse

```go
func NewResponse(req *Request, result interface{}, err error) (*Response, error)
```
NewResponse creates a new Response instance based on a Request.

#### func (*Response) MarshalJSON

```go
func (r *Response) MarshalJSON() ([]byte, error)
```

#### func (*Response) UnmarshalJSON

```go
func (r *Response) UnmarshalJSON(data []byte) error
```

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
func NewTracker(socketPath string) (*Tracker, error)
```
NewTracker creates and initializes a new Tracker. If a socketDir is not
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

#### func (*Tracker) NumRequests

```go
func (t *Tracker) NumRequests() int
```
NumRequests returns the number of tracked requests

#### func (*Tracker) ProxyUnix

```go
func (t *Tracker) ProxyUnix(req *Request) (*Request, error)
```
ProxyUnix proxies requests that have response hooks of non-unix sockets through
one that does. If the response hook is already a unix socket, it returns the
original request. If not, it tracks the original request and returns a new
request with a unix socket response hook. The purpose of this is so that there
can be a single entry and exit point for external communication, while local
services can reply directly to each other.

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
func (t *Tracker) TrackRequest(req *Request) error
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
func NewUnixListener(socketPath string) *UnixListener
```
NewUnixListener creates and initializes a new UnixListener.

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
func (ul *UnixListener) Stop()
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
