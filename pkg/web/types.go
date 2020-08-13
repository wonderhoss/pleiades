package web

import (
	"net/http"

	"github.com/gargath/pleiades/pkg/util"
	"github.com/go-redis/redis/v8"
)

// Frontend is the web frontend server
type Frontend struct {
	listenAddr string
	redis      *util.RedisOpts
	s          *http.Server
	r          *redis.Client
}

// Opts configure the frontend server
type Opts struct {
	Redis      *util.RedisOpts
	ListenAddr string
}

// Counters is the return type for the stats API
type Counters struct {
	Since    int64
	Counters []Counter
}

// Counter is a single redis counter value
type Counter struct {
	Name        string
	Description string
	Value       int64
}
