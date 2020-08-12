package file

import (
	"github.com/gargath/pleiades/pkg/aggregator"
	"github.com/gargath/pleiades/pkg/spinner"
	"github.com/go-redis/redis/v8"
)

// Aggregator is an aggregator implementation that reads from the filesystem
type Aggregator struct {
	File    *Opts
	stop    chan (bool)
	Redis   *aggregator.RedisOpts
	r       *redis.Client
	spinner *spinner.Spinner
}

// Opts hold config options for the file publisher
type Opts struct {
	Source string
}
