# health

[![health](https://godoc.org/github.com/cerana/cerana/providers/health?status.svg)](https://godoc.org/github.com/cerana/cerana/providers/health)



## Usage

#### type FileArgs

```go
type FileArgs struct {
	Path     string      `json:"path"`
	NotExist bool        `json:"notExist"`
	Mode     os.FileMode `json:"mode"`
	MinSize  int64       `json:"minSize"`
	MaxSize  int64       `json:"maxSize"`
}
```

FileArgs are arguments for the File health check.

#### type HTTPStatusArgs

```go
type HTTPStatusArgs struct {
	URL        string `json:"url"`
	Method     string `json:"method"`
	Body       []byte `json:"body"`
	StatusCode int    `json:"statusCode"`
}
```

HTTPStatusArgs are arguments for HTTPStatus health checks.

#### type Health

```go
type Health struct {
}
```

Health is a provider of health checks.

#### func  New

```go
func New(config *provider.Config, tracker *acomm.Tracker) *Health
```
New create a new instance of a Health provider.

#### func (*Health) File

```go
func (h *Health) File(req *acomm.Request) (interface{}, *url.URL, error)
```
File checks one or more attributes against supplied constraints.

#### func (*Health) HTTPStatus

```go
func (h *Health) HTTPStatus(req *acomm.Request) (interface{}, *url.URL, error)
```
HTTPStatus makes an HTTP request to the specified URL and compares the response
status code to an expected status code.

#### func (*Health) RegisterTasks

```go
func (h *Health) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of Health's task handlers with the server.

#### func (*Health) TCPResponse

```go
func (h *Health) TCPResponse(req *acomm.Request) (interface{}, *url.URL, error)
```
TCPResponse makes a TCP request to the specified address and checks the response
for a match to a specified string or regex.

#### func (*Health) Uptime

```go
func (h *Health) Uptime(req *acomm.Request) (interface{}, *url.URL, error)
```
Uptime checks a process's uptime against a minimum value.

#### type Mock

```go
type Mock struct {
	Data MockData
}
```

Mock is a mock Health provider.

#### func  NewMock

```go
func NewMock() *Mock
```
NewMock creates a new mock provider and initializes data.

#### func (*Mock) File

```go
func (m *Mock) File(req *acomm.Request) (interface{}, *url.URL, error)
```
File is a mock file health check.

#### func (*Mock) HTTPStatus

```go
func (m *Mock) HTTPStatus(req *acomm.Request) (interface{}, *url.URL, error)
```
HTTPStatus is a mock http status health check.

#### func (*Mock) RegisterTasks

```go
func (m *Mock) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of the Mock health task handlers with the server.

#### func (*Mock) TCPResponse

```go
func (m *Mock) TCPResponse(req *acomm.Request) (interface{}, *url.URL, error)
```
TCPResponse is a mock tcp response health check.

#### func (*Mock) Uptime

```go
func (m *Mock) Uptime(req *acomm.Request) (interface{}, *url.URL, error)
```
Uptime is a mock uptime health check.

#### type MockData

```go
type MockData struct {
	Uptime      bool
	File        bool
	TCPResponse bool
	HTTPStatus  bool
}
```

MockData is mock data for the Mock provider.

#### type TCPResponseArgs

```go
type TCPResponseArgs struct {
	Address string `json:"address"`
	Body    []byte `json:"body"`
	Regexp  string `json:"regexp"`
}
```

TCPResponseArgs ar arguments for TCPResponse health checks.

#### type UptimeArgs

```go
type UptimeArgs struct {
	Name      string        `json:"name"`
	MinUptime time.Duration `json:"minUptime"`
}
```

UptimeArgs are arguments for the uptime health check.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
