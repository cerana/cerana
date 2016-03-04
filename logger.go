package gozfs

var logger Logger = &defaultLogger{}

// Logger logs messages based on set logging level
type Logger interface {
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Panic(...interface{})
	Fatal(...interface{})
}

// defaultLogger is a no-op Logger
type defaultLogger struct{}

func (d *defaultLogger) Debug(args ...interface{}) {}
func (d *defaultLogger) Info(args ...interface{})  {}
func (d *defaultLogger) Warn(args ...interface{})  {}
func (d *defaultLogger) Error(args ...interface{}) {}
func (d *defaultLogger) Panic(args ...interface{}) {}
func (d *defaultLogger) Fatal(args ...interface{}) {}

// SetLogger sets the logger to log with
func SetLogger(l Logger) {
	if l == nil {
		l = &defaultLogger{}
	}
	logger = l
}
