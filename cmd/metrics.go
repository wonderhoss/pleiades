package main

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var metricsServer *http.Server

func initMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	metricsServer = &http.Server{Addr: ":9000"}
	logger.Debug("Starting metrics server")
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				logger.Errorf("Error from metrics server: %v", err)
			}
		}
	}()

}

func stopMetrics() {
	logger.Debug("Stopping metrics server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := metricsServer.Shutdown(ctx); err != nil {
		logger.Errorf("Error during metrics server shutdown: %v", err)
	}
}
