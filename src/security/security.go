package security

import (
	"crypto/aes"
	"crypto/cipher"
)

type Cipher struct{ aead cipher.AEAD }

func NewCipher(key []byte) (*Cipher, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCMWithRandomNonce(block)

	if err != nil {
		return nil, err
	}

	return &Cipher{aead}, nil
}

// Appends the encrypted data to dst. Pass a reused buffer as dst to avoid one allocation per chunk
func (c *Cipher) Encrypt(dst []byte, data []byte, additionalData []byte) []byte {
	return c.aead.Seal(dst, nil, data, additionalData)
}

// Appends the decrypted data to dst. Pass a reused buffer as dst to avoid one allocation per chunk
func (c *Cipher) Decrypt(dst []byte, data []byte, additionalData []byte) ([]byte, error) {
	return c.aead.Open(dst, nil, data, additionalData)
}
