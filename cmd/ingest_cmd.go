package main

import (
	"github.com/gargath/pleiades/pkg/ingester"
	"github.com/gargath/pleiades/pkg/ingester/publisher/file"
	"github.com/gargath/pleiades/pkg/ingester/publisher/kafka"
	"github.com/spf13/cobra"
)

var (
	cmdIngest = &cobra.Command{
		Use:   "ingest",
		Short: "Starts Pleiades ingest server",
		Long: `The ingest command starts the ingest server.
It will begin consuming the WMF stream and publish received events to the configured publisher.`,
		RunE: startIngest,
	}

	c      *ingester.Coordinator
	resume bool
)

func init() {
	cmdIngest.Flags().BoolVarP(&resume, "resume", "r", true, "try to resume from last seen event ID")
}

func startIngest(cmd *cobra.Command, args []string) error {

	logger.Info("Ingest server starting...")

	c = &ingester.Coordinator{
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

	registerShutdownHook(c)

	lastEventID, err := c.Start()
	if err != nil {
		return err
	}
	logger.Info("Ingest shutdown complete")
	logger.Infof("Last seen Event ID: %s", lastEventID)
	return nil
}
