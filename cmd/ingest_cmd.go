package main

import (
	"fmt"

	"github.com/gargath/pleiades/pkg/coordinator"
	"github.com/gargath/pleiades/pkg/log"
	"github.com/gargath/pleiades/pkg/publisher/file"
	"github.com/gargath/pleiades/pkg/publisher/kafka"
	"github.com/spf13/cobra"
)

var (
	cmdIngest = &cobra.Command{
		Use:   "ingest",
		Short: "Starts Pleiades ingest server",
		Long: `ingest starts the ingest server.
It will begin consuming the WMF stream and publish received events to the configured publisher.`,
		RunE: startIngest,
	}

	metricsPort string
	resume      bool
	fileOn      bool
	kafkaOn     bool
	fileDir     string
	kafkaBroker string
	kafkaTopic  string
)

func init() {
	cmdIngest.Flags().BoolVarP(&resume, "resume", "r", true, "try to resume from last seen event ID")
	cmdIngest.Flags().StringVar(&metricsPort, "metricsPort", "9000", "the port to serve Prometheus metrics on")
	cmdIngest.Flags().BoolVar(&fileOn, "file.enable", false, "enable the filesystem publisher")
	cmdIngest.Flags().BoolVar(&kafkaOn, "kafka.enable", false, "enable the kafka publisher")
	cmdIngest.Flags().StringVar(&fileDir, "file.publishDir", "./events", "the directory to publish events to")
	cmdIngest.Flags().StringVar(&kafkaBroker, "kafka.broker", "localhost:9092", "the kafka broker to connect to")
	cmdIngest.Flags().StringVar(&kafkaTopic, "kafka.topic", "pleiades-events", "the kafka topic to publish to")
}

func startIngest(cmd *cobra.Command, args []string) error {

	if fileOn && kafkaOn {
		return fmt.Errorf("Can only specify either --file.enable or --kafka.enable")

	} else if !fileOn && !kafkaOn {
		return fmt.Errorf("No publisher specified")
	}

	if verbose {
		log.InitLogLevel(log.VERBOSE)
	} else if quiet {
		log.InitLogLevel(log.QUIET)
	} else {
		log.InitLogLevel(log.DEFAULT)
	}
	logger.Info("Ingest starting up...")

	c = &coordinator.Coordinator{
		Resume: resume,
	}

	if fileOn {
		c.File = &file.Opts{
			Destination: fileDir,
		}
	}
	if kafkaOn {
		c.Kafka = &kafka.Opts{
			Broker: kafkaBroker,
			Topic:  kafkaTopic,
		}
	}

	registerShutdownHook()

	initMetrics(metricsPort)

	lastEventID, err := c.Start()
	if err != nil {
		return err
	}
	stopMetrics()
	logger.Info("Ingest shutdown complete")
	logger.Infof("Last seen Event ID: %s", lastEventID)
	return nil
}
