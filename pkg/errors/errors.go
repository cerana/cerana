package errors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// StackDepth is the maximum callstack depth that will be used.
const stackDepth uint = 32

type errorExt struct {
	cause   error
	context []string
	data    map[string]interface{}
	pcs     []uintptr
}

func (e *errorExt) Error() string {
	var buffer bytes.Buffer

	for i := len(e.context) - 1; i >= 0; i-- {
		_, _ = buffer.WriteString(e.context[i])
		_, _ = buffer.WriteString(": ")
	}

	_, _ = buffer.WriteString(e.cause.Error())

	return buffer.String()
}

func (e *errorExt) MarshalJSON() ([]byte, error) {
	// determine cause error's json exported fields and values
	causeMap := make(map[string]interface{})
	causeJSON, err := json.Marshal(e.cause)
	if err != nil {
		return nil, Wrap(err, "marshal error cause")
	}
	if err = json.Unmarshal(causeJSON, &causeMap); err != nil {
		return nil, Wrap(err, "unmarshal error cause")
	}

	// flatten caller data and cause error fields in output structure
	outputMap := make(map[string]interface{})
	for k, v := range causeMap {
		outputMap[k] = v
	}
	for k, v := range e.data {
		outputMap[k] = v
	}

	// add library fields last to avoid being overwritten
	outputMap["cause"] = e.Error()
	outputMap["stack"] = callstack(e.pcs)

	output, err := json.Marshal(outputMap)
	if err != nil {
		return nil, Wrap(err, "marshal error output map")
	}
	return output, nil
}

// Cause returns the original cause of the error.
func Cause(e error) error {
	if e == nil {
		return nil
	}

	if eExt, ok := e.(*errorExt); ok {
		return eExt.cause
	}

	return e
}

func fromError(err error) *errorExt {
	if eExt, ok := err.(*errorExt); ok {
		return eExt
	}

	// Skip 4 for getPCs:
	// 0 : runtime.Callers
	// 1 : errors.getPCs
	// 2 : errors.fromError
	// 3 : errors.{API}
	// 4 : {API caller}

	return &errorExt{
		cause:   err,
		context: make([]string, 0, 5),
		data:    make(map[string]interface{}),
		pcs:     getPCs(4),
	}
}

// New returns a new error with callstack that formats as the given text.
func New(msg string) error {
	return fromError(errors.New(msg))
}

// Newf returns a new error with callstack that formats according to a format
// specifier.
func Newf(format string, args ...interface{}) error {
	return fromError(fmt.Errorf(format, args...))
}

// Newv returns a new error with the callstack that formats as the given text
// and associates the supplied data with the error.
func Newv(msg string, values map[string]interface{}) error {
	return Wrapv(fromError(errors.New(msg)), values)
}

// Wrap wraps an error, creating a callstack if necessary and prefixing the
// underlying error message with the optional given text.
func Wrap(err error, msg ...string) error {
	if err == nil {
		return nil
	}

	eExt := fromError(err)
	if len(msg) > 0 {
		eExt.context = append(eExt.context, strings.Join(msg, " "))
	}

	return eExt
}

// Wrapf wraps an error, creating a callstack if necessary and prefixing the
// underlying error message with the given text formatted with the specifier.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	eExt := fromError(err)
	eExt.context = append(eExt.context, fmt.Sprintf(format, args...))

	return eExt
}

// Wrapv wraps an error, creating a callstack if necessary and associating the
// supplied data with the error.
func Wrapv(err error, values map[string]interface{}, msg ...string) error {
	if err == nil {
		return nil
	}

	eExt := Wrap(fromError(err), msg...).(*errorExt)
	for k, v := range values {
		eExt.data[k] = v
	}

	return eExt
}

func getPCs(skip int) []uintptr {
	pcs := make([]uintptr, stackDepth)
	n := runtime.Callers(skip, pcs)
	return pcs[0:n]
}

func callstack(pcs []uintptr) []string {
	stack := make([]string, len(pcs))
	for i, pc := range pcs {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		fnName := fn.Name()
		stack[i] = fmt.Sprintf("%s (%s:%d)", fnName, file, line)
	}
	return stack
}
