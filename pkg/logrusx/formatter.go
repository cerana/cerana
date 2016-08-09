// Package logrusx is a logrus formatter that adds better error value handling
// to the logrus.JSONFormatter
package logrusx

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
)

type (
	// JSONFormatter is a custom formatter extending logrus.JSONFormatter with better handling of error values
	JSONFormatter struct {
		logrus.JSONFormatter
	}

	// FieldError contains both the error struct and error message as explicit properties, including both when JSON marshaling.
	FieldError struct {
		Error   error
		Message string
		Stack   []string
	}
)

// Format replaces any error field values with a FieldError and produces a JSON formatted log entry
func (f *JSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	for k, v := range entry.Data {
		if err, ok := v.(error); ok {
			// Get the call stack and remove this function call from it
			stack := f.callStack()
			var stackStart int
			for stackStart = 1; stackStart < len(stack); stackStart++ {
				if !(strings.Contains(stack[stackStart], "github.com/Sirupsen/logrus") || strings.Contains(stack[stackStart], "github.com/cerana/cerana/pkg/logrusx")) {
					break
				}
			}

			entry.Data[k] = FieldError{
				Error:   err,
				Message: err.Error(),
				Stack:   stack[stackStart:],
			}
		}
	}
	return f.JSONFormatter.Format(entry)
}

func (f *JSONFormatter) callStack() []string {
	stack := make([]string, 0, 4)
	for i := 1; ; i++ {
		// TODO: use runtime.Callers && runtime.CallersFrames when go1.7 is out
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Look up the function name (package.FnName)
		fnName := runtime.FuncForPC(pc).Name()
		// Add the line to the stack, skipping anything from within the logrus
		// package so it starts at the log caller
		if !strings.HasPrefix(fnName, "github.com/Sirupsen/logrus.") {
			stack = append(stack, fmt.Sprintf("%s:%d (%s)", file, line, fnName))
		}
	}
	return stack
}
