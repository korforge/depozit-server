package main

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

func (c *Cipher) Encrypt(data []byte, path string) []byte {
	return c.aead.Seal(nil, nil, data, []byte(path))
}

func (c *Cipher) Decrypt(data []byte, path string) ([]byte, error) {
	return c.aead.Open(nil, nil, data, []byte(path))
}
