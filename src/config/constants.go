package config

const (
	DefaultPageSize            = 1000
	MaxPageSize                = 10000
	DefaultFilePermission      = 0600
	DefaultDirectoryPermission = 0700
	DefaultContainer           = "default"
)

// Plaintext size of every stored chunk except the last one of a file
const ChunkSize = 64 << 10

// ChunkSize plus the 12 byte nonce and 16 byte tag of AES-GCM
const MaxEncryptedChunkSize = ChunkSize + 28
