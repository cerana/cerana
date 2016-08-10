package logrusx

import (
	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/errors"
)

// DefaultSetup sets the logrus formatter and log level
func DefaultSetup(logLevel string) error {
	err := SetLevel(logLevel)
	if err != nil {
		return err
	}
	logrus.SetFormatter(&JSONFormatter{})
	return nil
}

// SetLevel parses and sets the log level
func SetLevel(logLevel string) error {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)
	return nil
}

// LogReturnedErr wraps a function that returns an error, calls the function,
// and logs any error.
// Useful for basic defer, e.g.
// `defer LogReturnedErr(f.Close,logrus.Fields{"file":f.Name()}, "failed to close file")`
func LogReturnedErr(fn func() error, fields map[string]interface{}, message string) {
	if err := fn(); err != nil {
		logrus.WithField("error", errors.Wrapv(err, fields)).Error(message)
	}
}

// DieOnError is a convenience method for checking an error, and if present,
// logging a Fatal message.
func DieOnError(err error, msg ...interface{}) {
	if err != nil {
		logrus.WithField("error", errors.Wrap(err)).Fatal(msg...)
	}
}
