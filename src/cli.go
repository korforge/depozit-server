package main

import (
	"depozit-server/src/config"
	"flag"
	"fmt"
	"os"
)

func CliLoadConfig() *config.Config {
	port := flag.String("port", "5051", "port to run gRPC server")
	defaultContainer := flag.String("default-container", "default", "Default Container")
	apiKey := flag.String("root-api-key", "", "API key for authentication")
	storagePath := flag.String("storage-path", "", "Storage path")

	flag.Parse()

	missing := false

	if *apiKey == "" {
		fmt.Println("Error: --root-api-key is required")
		missing = true
	}
	if *storagePath == "" {
		fmt.Println("Error: --storage-path is required")
		missing = true
	}

	if missing {
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}

	cfg := config.LoadConfig(*port, *apiKey, *defaultContainer, *storagePath)

	return cfg
}
