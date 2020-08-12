package frontend

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gargath/pleiades/pkg/log"
	"github.com/gargath/pleiades/pkg/web/static"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

const moduleName = "frontend"

var (
	logger = log.MustGetLogger(moduleName)
)

// Stop stops the frontend webserver
func (f *Frontend) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := f.s.Shutdown(ctx)
	if err != nil {
		logger.Errorf("Error shutting down: %v", err)
	}
}

// Start starts the server
func (f *Frontend) Start() error {
	r := mux.NewRouter()
	sr := r.PathPrefix("/api").Subrouter()
	sr.HandleFunc("/stats", f.statsHandler)
	//	s.HandleFunc("/stats/{key}", f.singleStatHandler)
	//	r.HandleFunc("/ws", f.websocketHandler)

	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(static.Assets)))

	s := &http.Server{
		Addr:    f.listenAddr,
		Handler: r,
	}
	f.s = s
	logger.Infof("Frontend server started")
	return f.s.ListenAndServe()
}

// NewFrontend initialized a frontend server
func NewFrontend(fo *Opts) (*Frontend, error) {
	r := redis.NewClient(&redis.Options{
		Addr: fo.Redis.RedisAddr,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pong, err := r.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %v", fo.Redis.RedisAddr, err)
	}
	logger.Debugf("Connected to Redis: %v", pong)

	s := &Frontend{
		redis:      fo.Redis,
		listenAddr: fo.ListenAddr,
	}
	s.r = r
	return s, nil
}
