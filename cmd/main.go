package main

import (
	"fmt"
	"log"
	"os"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/gargath/template/pkg/placeholder"
)

func main() {
	log.Printf("Placeholder %s\n", version())

	viper.SetEnvPrefix("PLACEHOLDER")
	viper.AutomaticEnv()

	//	flag.String("listenAddr", "0.0.0.0:8080", "address to listen on")
	flag.Bool("help", false, "print this help and exit")

	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	if viper.GetBool("help") {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, flag.CommandLine.FlagUsages())
		os.Exit(0)
	}

	fmt.Printf("%s\n", placeholder.Hello("foo"))
}
