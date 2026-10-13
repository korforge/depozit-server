package config

import (
	"os"
)

type Config struct {
	Host                string
	Port                string
	APIKey              string
	StoragePath         string
	EncryptionKey       []byte
	Permission          os.FileMode
	DirectoryPermission os.FileMode
}

func LoadConfig(host string, port string, rootApiKey string, storagePath string, encryptionKey []byte) *Config {
	return &Config{
		Host:                host,
		Port:                port,
		APIKey:              rootApiKey,
		EncryptionKey:       encryptionKey,
		StoragePath:         storagePath,
		Permission:          0600,
		DirectoryPermission: 0700,
	}
}
