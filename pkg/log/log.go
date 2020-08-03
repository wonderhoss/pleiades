package log

import (
	"os"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

func init() {
	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{module:-10s}| %{shortfunc:-20s}	▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	//logging.SetBackend(backend, backendFormatter)
	logging.SetBackend(backendFormatter)
}

// MustGetLogger returns a logger for the given package or panics
func MustGetLogger(moduleName string) *logging.Logger {
	initLogLevel(moduleName)
	return logging.MustGetLogger(moduleName)
}

func initLogLevel(moduleName string) {
	var verbose logging.Level
	if viper.GetBool("verbose") {
		verbose = logging.DEBUG
	} else if viper.GetBool("quiet") {
		verbose = logging.ERROR
	} else {
		verbose = logging.INFO
	}
	logging.SetLevel(verbose, moduleName)
}
