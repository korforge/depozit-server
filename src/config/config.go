package config

import (
	"os"
)

type Config struct {
	Port             string
	APIKey           string
	DefaultContainer string
	StoragePath      string
	Permission       os.FileMode
}

func LoadConfig(Port string, rootApiKey string, defaultContainer string, storagePath string) *Config {
	return &Config{
		Port:             Port,
		APIKey:           rootApiKey,
		DefaultContainer: defaultContainer,
		StoragePath:      storagePath,
		Permission:       0755,
	}
}
