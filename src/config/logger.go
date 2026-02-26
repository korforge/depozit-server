package config

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func SetupLogger() {
	handler := slog.NewTextHandler(os.Stdout, nil) // writes to console
	Logger = slog.New(handler)
}
