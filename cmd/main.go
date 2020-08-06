package main

import (
	"fmt"
	"strings"

	//"log"

	"os"

	"github.com/op/go-logging"

	"github.com/spf13/cobra"

	"github.com/gargath/pleiades/pkg/coordinator"
)

const moduleName = "main"

var (
	c       *coordinator.Coordinator
	logger  *logging.Logger
	verbose bool
	quiet   bool
)

func main() {
	var cmdEcho = &cobra.Command{
		Use:   "echo [string to echo]",
		Short: "Echo anything to the screen",
		Long: `echo is for echoing anything back.
	Echo works a lot like print, except it has a child command.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Echo: " + strings.Join(args, " "))
		},
	}
	var rootCmd = &cobra.Command{Use: "pleiades"}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress all output except for errors")

	rootCmd.AddCommand(cmdEcho)
	rootCmd.AddCommand(cmdIngest)
	err := rootCmd.Execute()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(55)
	}
}
