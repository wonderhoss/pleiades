package main

import (
	"github.com/gargath/pleiades/pkg/aggregator"
	"github.com/gargath/pleiades/pkg/aggregator/file"
	"github.com/gargath/pleiades/pkg/aggregator/kafka"
	"github.com/gargath/pleiades/pkg/util"

	"github.com/spf13/cobra"
)

var (
	cmdAgg = &cobra.Command{
		Use:   "aggregate",
		Short: "Starts Pleiades stats aggregator",
		Long: `The aggregate command starts the stats aggregation server.
	It will consume events from kafka and write aggregate stats to redis.`,
		RunE: startAggregator,
	}

	redis            string
	redisUseSentinel bool
)

func init() { //TODO: Use Sentinels
	cmdAgg.Flags().StringVar(&redis, "redis-addr", "localhost:6379", "the Redis server to write aggregated stats to")
	cmdAgg.Flags().BoolVar(&redisUseSentinel, "redis-use-sentinel", false, "should Redis use Sentinel for connect")
}

func startAggregator(cmd *cobra.Command, args []string) error {
	logger.Info("Aggregation server starting...")

	var a aggregator.Server
	var aggErr error
	redisOpts := &util.RedisOpts{RedisAddr: redis, RedisUseSentinel: redisUseSentinel}
	if fileOn {
		a, aggErr = file.NewAggregator(redisOpts, &file.Opts{
			Source: fileDir,
		})
	}
	if kafkaOn {
		a, aggErr = kafka.NewAggregator(redisOpts, &kafka.Opts{
			Broker: kafkaBroker,
			Topic:  kafkaTopic,
		})
	}
	if aggErr != nil {
		return aggErr
	}

	registerShutdownHook(a)

	err := a.Start()
	if err != nil {
		return err
	}
	logger.Info("Aggregation shutdown complete")
	return nil
}
