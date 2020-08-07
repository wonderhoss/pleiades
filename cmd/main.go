package main

import (
	"fmt"
	"os"

	//"log"

	"github.com/op/go-logging"

	"github.com/spf13/cobra"

	"github.com/gargath/pleiades/pkg/log"
)

const moduleName = "main"

var (
	logger      *logging.Logger
	verbose     bool
	quiet       bool
	metricsPort string
	fileOn      bool
	kafkaOn     bool
	fileDir     string
	kafkaBroker string
	kafkaTopic  string
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "pleiades",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if verbose && quiet {
				return fmt.Errorf(" -quiet and -verbose are mutually exclusive")
			}
			if verbose {
				log.InitLogLevel(log.VERBOSE)
			} else if quiet {
				log.InitLogLevel(log.QUIET)
			} else {
				log.InitLogLevel(log.DEFAULT)
			}

			if fileOn && kafkaOn {
				return fmt.Errorf("Can only specify either --file.enable or --kafka.enable")

			} else if !fileOn && !kafkaOn {
				return fmt.Errorf("No queue backend specified (use either --file.enable or --kafka.enable)")
			}
			initMetrics(metricsPort)
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			stopMetrics()
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress all output except for errors")
	rootCmd.PersistentFlags().StringVar(&metricsPort, "metricsPort", "9000", "the port to serve Prometheus metrics on")
	rootCmd.PersistentFlags().BoolVar(&fileOn, "file.enable", false, "enable the filesystem publisher")
	rootCmd.PersistentFlags().StringVar(&fileDir, "file.publishDir", "./events", "the directory to publish events to")
	rootCmd.PersistentFlags().BoolVar(&kafkaOn, "kafka.enable", false, "enable the kafka publisher")
	rootCmd.PersistentFlags().StringVar(&kafkaBroker, "kafka.broker", "localhost:9092", "the kafka broker to connect to")
	rootCmd.PersistentFlags().StringVar(&kafkaTopic, "kafka.topic", "pleiades-events", "the kafka topic to publish to")

	rootCmd.AddCommand(cmdIngest)
	rootCmd.AddCommand(cmdAgg)

	logger = log.MustGetLogger(moduleName)
	logger.Infof("Pleiades %s\n", version())

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
