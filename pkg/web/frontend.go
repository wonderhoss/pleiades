package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gargath/pleiades/pkg/log"
	"github.com/gargath/pleiades/pkg/util"
	"github.com/gargath/pleiades/pkg/web/static"
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
	r.Use(prometheusMiddleware)
	sr := r.PathPrefix("/api").Subrouter()
	sr.HandleFunc("/stats", f.statsHandler)
	sr.HandleFunc("/stats/{day}", f.statsForDayHandler)
	sr.HandleFunc("/days", f.daysHandler)
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
	r, err := util.NewValidatedRedisClient(fo.Redis)
	if err != nil {
		return nil, fmt.Errorf("Failed to create frontend: %v", err)
	}

	s := &Frontend{
		redis:      fo.Redis,
		listenAddr: fo.ListenAddr,
	}
	s.r = r
	return s, nil
}
