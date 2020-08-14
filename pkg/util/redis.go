package util

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/op/go-logging"
)

const moduleName = "util"

var (
	logger = logging.MustGetLogger(moduleName)
)

// NewValidatedRedisClient creates a new redis client and performs a PING before returning it
func NewValidatedRedisClient(opts *RedisOpts) (*redis.Client, error) {
	var r *redis.Client
	if opts.RedisUseSentinel {
		r = redis.NewFailoverClient(&redis.FailoverOptions{
			SentinelAddrs: []string{opts.RedisAddr},
			MasterName:    mymaster,
		})
	} else {
		r = redis.NewClient(&redis.Options{
			Addr: opts.RedisAddr,
		})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pong, err := r.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %v", opts.RedisAddr, err)
	}
	logger.Debugf("Connected to Redis: %v", pong)
	return r, nil
}
