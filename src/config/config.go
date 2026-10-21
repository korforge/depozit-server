package config

type Config struct {
	Host                 string
	Port                 string
	APIKey               string
	StoragePath          string
	EncryptionKey        []byte
	MaxFileSize          int64
	MaxConcurrentStreams uint32
	TLSCert              string
	TLSKey               string
}

func LoadConfig(host string, port string, rootApiKey string, storagePath string, encryptionKey []byte) *Config {
	return &Config{
		Host:          host,
		Port:          port,
		APIKey:        rootApiKey,
		EncryptionKey: encryptionKey,
		StoragePath:   storagePath,
	}
}
