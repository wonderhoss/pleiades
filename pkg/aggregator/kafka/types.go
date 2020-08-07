package kafka

import "github.com/gargath/pleiades/pkg/aggregator"

// Aggregator is an aggregator implementation that reads from the filesystem
type Aggregator struct {
	Kafka *Opts
	stop  chan (bool)
	Redis *aggregator.RedisOpts
}

// Opts hold configuration for the kafka publisheru
type Opts struct {
	Broker string
	Topic  string
}
