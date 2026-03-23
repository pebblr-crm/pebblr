// Command pebblr is the unified CLI for the Pebblr CRM.
//
// Usage:
//
//	pebblr serve   --config <path>    Start the API server
//	pebblr config  validate --config <path>  Validate a tenant config file
package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 2
	}

	switch args[0] {
	case "serve":
		return runServe(args[1:])
	case "config":
		return runConfig(args[1:])
	case "--help", "-h", "help":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printUsage()
		return 2
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  pebblr serve   --config <path>             Start the API server
  pebblr config  validate --config <path>    Validate a tenant config file`)
}
