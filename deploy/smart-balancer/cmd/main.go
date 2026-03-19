package main

import (
	"balancer/internal"
	"balancer/internal/config-parser/app"
	"log"

	"github.com/spf13/pflag"
)

func main() {
	configPath := pflag.StringP("config", "c", "config.yaml", "Path to application config file")
	pflag.Parse()

	config, err := app.LoadApplicationConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	internal.Run(config)
}
