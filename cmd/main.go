package main

import (
	"fmt"

	//"log"

	"os"
	"os/signal"

	"github.com/op/go-logging"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/gargath/pleiades/pkg/coordinator"
	"github.com/gargath/pleiades/pkg/log"
	"github.com/gargath/pleiades/pkg/publisher/file"
	"github.com/gargath/pleiades/pkg/publisher/kafka"
)

const moduleName = "main"

var (
	c      *coordinator.Coordinator
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
	if !viper.GetBool("kafka.enable") && !viper.GetBool("file.enable") {
		fmt.Fprintf(os.Stderr, "ERROR: no publisher enabled\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, flag.CommandLine.FlagUsages())
		os.Exit(0)
	}
	if viper.GetBool("kafka.enable") && viper.GetBool("file.enable") {
		fmt.Fprintf(os.Stderr, "ERROR: only one publisher can be enabled\n")
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
	flag.BoolP("quiet", "q", false, "quiet output - only show ERROR and above")
	flag.BoolP("resume", "r", true, "try to resume from last seen event ID")

	flag.Bool("file.enable", false, "enable the file publisher")
	flag.String("file.publishDir", "./events", "the directory to publish events to")

	flag.Bool("kafka.enable", false, "enable the kafka publisher")
	flag.String("kafka.broker", "localhost:9092", "the kafka broker to connect to")
	flag.String("kafka.topic", "pleiades-events", "the kafka topic to publish to")

	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	validateFlags()

	logger = log.MustGetLogger(moduleName)
	if viper.GetBool("verbose") {
		log.InitLogLevel(log.VERBOSE)
	} else if viper.GetBool("quiet") {
		log.InitLogLevel(log.QUIET)
	} else {
		log.InitLogLevel(log.DEFAULT)
	}
	logger.Infof("Pleiades %s\n", version())

	c = &coordinator.Coordinator{
		Resume: viper.GetBool("resume"),
	}

	if viper.GetBool("file.enable") && viper.GetBool("kafka.enable") {
		logger.Error("Can only specify either --file.enable or --kafka.enable")
		os.Exit(1)
	} else if !viper.GetBool("file.enable") && !viper.GetBool("kafka.enable") {
		logger.Error("No publisher specified")
		os.Exit(1)
	}

	if viper.GetBool("file.enable") {
		c.File = &file.Opts{
			Destination: viper.GetString("file.publishDir"),
		}
	}
	if viper.GetBool("kafka.enable") {
		c.Kafka = &kafka.Opts{
			Broker: viper.GetString("kafka.broker"),
			Topic:  viper.GetString("kafka.topic"),
		}
	}

	registerShutdownHook()

	initMetrics(viper.GetString("metricsPort"))

	logger.Info("Starting up...")
	lastEventID, err := c.Start()
	if err != nil {
		logger.Errorf("Event consumer exited with error: %v", err)
	}
	stopMetrics()
	logger.Info("Shutdown complete")
	logger.Infof("Last seen Event ID: %s", lastEventID)
}
