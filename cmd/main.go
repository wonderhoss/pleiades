package main

import (
	"fmt"
	"os"

	//"log"

	"github.com/op/go-logging"

	"github.com/spf13/cobra"

	"github.com/gargath/pleiades/pkg/coordinator"
	"github.com/gargath/pleiades/pkg/log"
)

const moduleName = "main"

var (
	c       *coordinator.Coordinator
	logger  *logging.Logger
	verbose bool
	quiet   bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "pleiades",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if verbose && quiet {
				return fmt.Errorf(" -quiet and -verbose are mutually exclusive")
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress all output except for errors")

	rootCmd.AddCommand(cmdIngest)

	logger = log.MustGetLogger(moduleName)
	logger.Infof("Pleiades %s\n", version())

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
