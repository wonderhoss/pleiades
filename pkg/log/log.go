package log

import (
	"os"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
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
func InitLogLevel() {
	var verbose logging.Level
	if viper.GetBool("verbose") {
		verbose = logging.DEBUG
	} else if viper.GetBool("quiet") {
		verbose = logging.ERROR
	} else {
		verbose = logging.INFO
	}
	for _, m := range modules {
		logging.SetLevel(verbose, m)
	}
}
