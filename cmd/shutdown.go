package main

import (
	"os"
	"os/signal"
)

func registerShutdownHook(s Stoppable) {
	logger.Debug("Registering shutdown handler")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		logger.Debug("Shutting down...")
		s.Stop()
	}()
}
