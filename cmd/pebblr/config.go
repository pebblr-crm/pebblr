package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pebblr/pebblr/internal/config"
)

func runConfig(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: pebblr config validate --config <path>")
		return 2
	}

	switch args[0] {
	case "validate":
		return runConfigValidate(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown config subcommand: %s\n", args[0])
		return 2
	}
}

func runConfigValidate(args []string) int {
	fs := flag.NewFlagSet("config validate", flag.ContinueOnError)
	configPath := fs.String("config", defaultConfigPath, "path to tenant config JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	_, validationErrors, err := config.LoadAndValidate(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return 2
	}

	if len(validationErrors) > 0 {
		for _, e := range validationErrors {
			fmt.Fprintln(os.Stderr, e)
		}
		return 1
	}

	fmt.Println("ok")
	return 0
}
