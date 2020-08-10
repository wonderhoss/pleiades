package kafka

import (
	"github.com/gargath/pleiades/pkg/aggregator"
	"github.com/gargath/pleiades/pkg/spinner"
	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
)

// Aggregator is an aggregator implementation that reads from the filesystem
type Aggregator struct {
	Kafka   *Opts
	stop    chan (bool)
	Redis   *aggregator.RedisOpts
	r       *redis.Client
	k       *kafka.Reader
	spinner *spinner.Spinner
}

// Opts hold configuration for the kafka publisheru
type Opts struct {
	Broker string
	Topic  string
}
