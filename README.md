# acomm

[![acomm](https://godoc.org/github.com/mistifyio/async-comm?status.png)](https://godoc.org/github.com/mistifyio/async-comm)

Package acomm is a library for asynchronous communication between services.

## Usage

#### type Request

```go
type Request struct {
	ID           string      `json:"id"`
	ResponseHook *url.URL    `json:"responsehook"`
	Args         interface{} `json:"args"`
}
```

Request is a request data structure for asynchronous requests. The ID is used to
identify the request throught its life cycle. The ResponseHook is a URL where
response data should be sent.

#### func  NewRequest

```go
func NewRequest(responseHook string, args interface{}) (*Request, error)
```
NewRequest creates a new Request instance.

#### func (*Request) Respond

```go
func (req *Request) Respond(resp *Response) error
```
Respond sends a Response to a Request's ResponseHook.

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

#### func (*Response) Send

```go
func (resp *Response) Send(responseHook *url.URL) error
```
Send attempts send the Response to the specified URL.

#### type Tracker

```go
type Tracker struct {
}
```

Tracker keeps track of requests waiting on a response.

#### func  NewTracker

```go
func NewTracker(socketDir string) *Tracker
```
NewTracker creates and initializes a new Tracker. If a socketDir is not
provided, the response socket will be created in a temporary directory.

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

#### func (*Tracker) RetrieveRequest

```go
func (t *Tracker) RetrieveRequest(id string) *Request
```
RetrieveRequest returns a tracked Request based on ID and stops tracking it.
This should only be called directly when the Tracker is being used by an
original source of requests.

#### func (*Tracker) StartListener

```go
func (t *Tracker) StartListener() error
```
StartListener activates the tracker and starts listening for responses.

#### func (*Tracker) StopListener

```go
func (t *Tracker) StopListener(timeout time.Duration) error
```
StopListener disallows new requests to be tracked and waits until either all
active requests are handled or a timeout occurs. The chan returned will be used
to notify when the Tracker is fully stopped.

#### func (*Tracker) TrackRequest

```go
func (t *Tracker) TrackRequest(req *Request)
```
TrackRequest tracks a request. This should only be called directly when the
Tracker is being used by an original source of requests. Responses should then
be removed with RetrieveRequest.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
