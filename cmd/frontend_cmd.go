package main

import (
	"fmt"
	"net/http"

	"github.com/gargath/pleiades/pkg/frontend"

	"github.com/spf13/cobra"
)

var (
	cmdFront = &cobra.Command{
		Use:   "frontend",
		Short: "Starts Pleiades frontend server",
		Long: `The frontend command starts the frontend web server.
	It will serve a page displaying redis counters.`,
		RunE: startFrontend,
	}

	frontendRedis string
	listenAddr    string
)

func init() { //TODO: Use Sentinels
	cmdFront.Flags().StringVar(&frontendRedis, "frontend-redis-addr", "localhost:6379", "the Redis server to write aggregated stats to")
	cmdFront.Flags().StringVar(&listenAddr, "listen-addr", ":8080", "the address to listen on")

}

func startFrontend(cmd *cobra.Command, args []string) error {
	f, err := frontend.NewFrontend(&frontend.Opts{
		ListenAddr: listenAddr,
		Redis: &frontend.RedisOpts{
			RedisAddr: frontendRedis,
		},
	})

	if err != nil {
		return fmt.Errorf("Failed to start frontend server: %v", err)
	}

	registerShutdownHook(f)

	err = f.Start()
	logger.Info("Web server started")
	if (err != nil) && (err != http.ErrServerClosed) {
		return err
	}
	logger.Info("Web server shutdown complete")
	return nil
}