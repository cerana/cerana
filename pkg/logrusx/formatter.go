// Package logrusx is a logrus formatter that adds better error value handling
// to the logrus.JSONFormatter
package logrusx

import (
	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/errors"
)

type (
	// JSONFormatter is a custom formatter extending logrus.JSONFormatter with better handling of error values
	JSONFormatter struct {
		logrus.JSONFormatter
	}
)

// Format wraps the logrus.JSONFormatter.Format to pre-marshal wrapped errors
// rather than simply use the error message.
func (f *JSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	for k, v := range entry.Data {
		if err, ok := v.(error); ok {
			// Get the call stack and remove this function call from it
			j, e := json.Marshal(errors.Wrap(err))
			if e != nil {
				return nil, errors.Wrap(e)
			}
			raw := json.RawMessage(j)
			entry.Data[k] = &raw
		}
	}
	return f.JSONFormatter.Format(entry)
}
