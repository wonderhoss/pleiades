package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	cmdAgg = &cobra.Command{
		Use:   "aggregate",
		Short: "Starts Pleiades stats aggregator",
		Long: `aggregate starts the stats aggregation server.
	It will consume events from kafka and write aggregate stats to redis.`,
		RunE: startAggregator,
	}
)

func startAggregator(cmd *cobra.Command, args []string) error {
	logger.Info("Aggregation server starting...")
	return fmt.Errorf("Not implemented")
}
