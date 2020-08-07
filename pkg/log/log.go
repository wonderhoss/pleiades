package log

import (
	"os"

	"github.com/op/go-logging"
)

// Level defines available log levels
type Level int

const (
	// QUIET will suppress all log messages except error and above
	QUIET Level = 0
	// DEFAULT is info-level logging
	DEFAULT = 1
	// VERBOSE includes all debug logs
	VERBOSE = 2
)

var modules = []string{}

func init() {
	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{module:-15s}| %{shortfunc:-20s}	▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	//logging.SetBackend(backend, backendFormatter)
	logging.SetBackend(backendFormatter)
}

// MustGetLogger returns a logger for the given package or panics
func MustGetLogger(moduleName string) *logging.Logger {
	modules = append(modules, moduleName)
	return logging.MustGetLogger(moduleName)
}

// InitLogLevel sets all registered loggers to use the verbosity indicated as command line flags
func InitLogLevel(l Level) {
	var logl logging.Level
	switch l {
	case QUIET:
		logl = logging.ERROR
	case DEFAULT:
		logl = logging.INFO
	case VERBOSE:
		logl = logging.DEBUG
	}

	for _, m := range modules {
		logging.SetLevel(logl, m)
	}
}
