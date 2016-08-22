# logrusx

[![logrusx](https://godoc.org/github.com/cerana/cerana/pkg/logrusx?status.svg)](https://godoc.org/github.com/cerana/cerana/pkg/logrusx)

Package logrusx is a logrus formatter that adds better error value handling to
the logrus.JSONFormatter

## Usage

#### func  DefaultSetup

```go
func DefaultSetup(logLevel string) error
```
DefaultSetup sets the logrus formatter and log level

#### func  DieOnError

```go
func DieOnError(err error, msg ...interface{})
```
DieOnError is a convenience method for checking an error, and if present,
logging a Fatal message.

#### func  LogReturnedErr

```go
func LogReturnedErr(fn func() error, fields map[string]interface{}, message string)
```
LogReturnedErr wraps a function that returns an error, calls the function, and
logs any error. Useful for basic defer, e.g. `defer
LogReturnedErr(f.Close,logrus.Fields{"file":f.Name()}, "failed to close file")`

#### func  SetLevel

```go
func SetLevel(logLevel string) error
```
SetLevel parses and sets the log level

#### type JSONFormatter

```go
type JSONFormatter struct {
	logrus.JSONFormatter
}
```

JSONFormatter is a custom formatter extending logrus.JSONFormatter with better
handling of error values

#### func (*JSONFormatter) Format

```go
func (f *JSONFormatter) Format(entry *logrus.Entry) ([]byte, error)
```
Format wraps the logrus.JSONFormatter.Format to pre-marshal wrapped errors
rather than simply use the error message.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
