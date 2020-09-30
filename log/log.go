// Package log provides a log interface
package log

var (
	// Default logger
	DefaultLogger Logger
)

// Logger is a generic logging interface
type Logger interface {
	// Init initialises options
	Init(options ...Option) error
	// The Logger options
	Options() Options
	// Fields set fields to always be logged
	Fields(fields map[string]interface{}) *Helper
	// Log writes a log entry
	Log(level Level, msg string, v ...interface{})
	// Logf writes a formatted log entry
	Logf(level Level, format string, v ...interface{})
	// Close close log
	Close()
	// String returns the name of logger
	String() string
}

func Init(opts ...Option) error {
	return DefaultLogger.Init(opts...)
}

func Fields(fields map[string]interface{}) *Helper {
	return DefaultLogger.Fields(fields)
}

func Close() {
	DefaultLogger.Close()
}

func String() string {
	return DefaultLogger.String()
}
