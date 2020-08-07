package main

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var metricsServer *http.Server

func initMetrics(metricsPort string) {
	http.Handle("/metrics", promhttp.Handler())
	port := ":" + metricsPort
	metricsServer = &http.Server{Addr: port}
	logger.Debug("Starting metrics server")
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				logger.Fatalf("Error starting metrics server: %v", err)
				panic(err)
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
