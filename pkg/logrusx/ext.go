package logrusx

import (
	log "github.com/Sirupsen/logrus"
)

// DefaultSetup sets the logrus formatter and log level
func DefaultSetup(logLevel string) error {
	err := SetLevel(logLevel)
	if err != nil {
		return err
	}
	log.SetFormatter(&MistifyFormatter{})
	return nil
}

// SetLevel parses and sets the log level
func SetLevel(logLevel string) error {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	log.SetLevel(level)
	return nil
}

// LogReturnedErr wraps a function that returns an error, calls the function,
// and logs any error.
// Useful for basic defer, e.g.
// `defer LogReturnedErr(f.Close(),logrus.Fields{"file":f.Name()}, "failed to close file")`
func LogReturnedErr(fn func() error, fields log.Fields, message string) {
	if err := fn(); err != nil {
		if fields == nil {
			fields = log.Fields{}
		}
		fields["error"] = err
		log.WithFields(fields).Error(message)
	}
}
