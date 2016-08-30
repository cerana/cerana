# errors

[![errors](https://godoc.org/github.com/cerana/cerana/pkg/errors?status.svg)](https://godoc.org/github.com/cerana/cerana/pkg/errors)

Package errors provides richer error handling functionality

Errors created or wrapped include a stack trace, optional additional context
messages, and optional additional relevant data.

## Usage

#### func  Cause

```go
func Cause(e error) error
```
Cause returns the original cause of the error.

#### func  New

```go
func New(msg string) error
```
New returns a new error with callstack that formats as the given text.

#### func  Newf

```go
func Newf(format string, args ...interface{}) error
```
Newf returns a new error with callstack that formats according to a format
specifier.

#### func  Newv

```go
func Newv(msg string, values map[string]interface{}) error
```
Newv returns a new error with the callstack that formats as the given text and
associates the supplied data with the error.

#### func  ResetStack

```go
func ResetStack(err error) error
```
ResetStack generates a new stack trace for the error at the current location if
one was present.

#### func  Wrap

```go
func Wrap(err error, msg ...string) error
```
Wrap wraps an error, creating a callstack if necessary and prefixing the
underlying error message with the optional given text.

#### func  Wrapf

```go
func Wrapf(err error, format string, args ...interface{}) error
```
Wrapf wraps an error, creating a callstack if necessary and prefixing the
underlying error message with the given text formatted with the specifier.

#### func  Wrapv

```go
func Wrapv(err error, values map[string]interface{}, msg ...string) error
```
Wrapv wraps an error, creating a callstack if necessary and associating the
supplied data with the error.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
