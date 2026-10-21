package main

import (
	"depozit-server/src/config"
	"depozit-server/src/server"
)

func main() {
	config.SetupLogger()
	cfg := CliLoadConfig()
	server.RunServer(cfg)
}
