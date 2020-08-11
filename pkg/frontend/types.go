package frontend

import (
	"net/http"

	"github.com/go-redis/redis/v8"
)

// Frontend is the web frontend server
type Frontend struct {
	listenAddr string
	redis      *RedisOpts
	s          *http.Server
	r          *redis.Client
}

// RedisOpts contains redis configuration
type RedisOpts struct {
	RedisAddr string
}

// Opts configure the frontend server
type Opts struct {
	Redis      *RedisOpts
	ListenAddr string
}

// Counters is the return type for the stats API
type Counters struct {
	Counters []Counter
}

// Counter is a single redis counter value
type Counter struct {
	Name        string
	Description string
	Value       int64
}
