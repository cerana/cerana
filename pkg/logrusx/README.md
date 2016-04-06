# logrusx

[![logrusx](https://godoc.org/github.com/cerana/cerana/pkg/logrusx?status.png)](https://godoc.org/github.com/cerana/cerana/pkg/logrusx)

Package logrusx is a logrus formatter that adds better error value handling to
the logrus.JSONFormatter

## Usage

#### func  DefaultSetup

```go
func DefaultSetup(logLevel string) error
```
DefaultSetup sets the logrus formatter and log level

#### func  LogReturnedErr

```go
func LogReturnedErr(fn func() error, fields log.Fields, message string)
```
LogReturnedErr wraps a function that returns an error, calls the function, and
logs any error. Useful for basic defer, e.g. `defer
LogReturnedErr(f.Close(),logrus.Fields{"file":f.Name()}, "failed to close
file")`

#### func  SetLevel

```go
func SetLevel(logLevel string) error
```
SetLevel parses and sets the log level

#### type FieldError

```go
type FieldError struct {
	Error   error
	Message string
	Stack   []string
}
```

FieldError contains both the error struct and error message as explicit
properties, including both when JSON marshaling.

#### type MistifyFormatter

```go
type MistifyFormatter struct {
	log.JSONFormatter
}
```

MistifyFormatter is a custom logrus formatter extending JSONFormatter

#### func (*MistifyFormatter) Format

```go
func (f *MistifyFormatter) Format(entry *log.Entry) ([]byte, error)
```
Format replaces any error field values with a FieldError and produces a JSON
formatted log entry

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
