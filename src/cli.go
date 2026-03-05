package main

import (
	"depozit-server/src/config"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func CliLoadConfig() *config.Config {

	envPath := flag.String("env", ".env", "Path to .env file with parameters")

	flag.Parse()

	//defaultContainer := flag.String("default-container", "default", "Default Container")
	//apiKey := flag.String("root-api-key", "", "API key for authentication")
	//storagePath := flag.String("storage-path", "", "Storage path")

	err := godotenv.Load(*envPath)

	if err != nil {
		fmt.Println("Error loading .env file")
	}

	port := os.Getenv("PORT")
	apiKey := os.Getenv("ROOT_API_KEY")
	storagePath := os.Getenv("STORAGE_PATH")
	defaultContainer := os.Getenv("CONTAINER")

	missing := false

	if apiKey == "" {
		fmt.Println("Error: --root-api-key is required")
		missing = true
	}
	if storagePath == "" {
		fmt.Println("Error: --storage-path is required")
		missing = true
	}

	if missing {
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}

	cfg := config.LoadConfig(port, apiKey, defaultContainer, storagePath)

	return cfg
}
