package main

import (
	"fmt"

	//"log"

	"os"
	"os/signal"

	"github.com/op/go-logging"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/gargath/pleiades/pkg/consumer"
	"github.com/gargath/pleiades/pkg/log"
)

const moduleName = "main"

var (
	c      *consumer.Consumer
	logger *logging.Logger
)

func registerShutdownHook() {
	logger.Debug("Registering shutdown handler")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		logger.Debug("Shutting down...")
		c.Stop()
	}()
}

func validateFlags() {
	if viper.GetBool("help") {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, flag.CommandLine.FlagUsages())
		os.Exit(0)
	}
	if viper.GetBool("verbose") && viper.GetBool("quiet") {
		fmt.Fprintf(os.Stderr, "ERROR: -quiet and -verbose are mutually exclusive\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, flag.CommandLine.FlagUsages())
		os.Exit(0)
	}
}

func main() {

	viper.SetEnvPrefix("PLEIADES")
	viper.AutomaticEnv()

	//	flag.String("listenAddr", "0.0.0.0:8080", "address to listen on")
	flag.Bool("help", false, "print this help and exit")
	flag.String("metricsPort", "9000", "the port to serve Prometheus metrics on")
	flag.BoolP("verbose", "v", false, "enable verbose output")
	flag.BoolP("quiet", "q", false, "quiet output")

	flag.Parse()
	viper.BindPFlags(flag.CommandLine)
	logger = log.MustGetLogger(moduleName)
	logger.Infof("Pleiades %s\n", version())

	validateFlags()

	c = &consumer.Consumer{}

	registerShutdownHook()

	initMetrics()

	logger.Info("Starting to consume events")
	lastEventID, err := c.Start()
	if err != nil {
		logger.Errorf("Event consumer exited with error: %v", err)
	}
	stopMetrics()
	logger.Info("Shutdown complete")
	logger.Infof("Last seen Event ID: %s", lastEventID)
}
