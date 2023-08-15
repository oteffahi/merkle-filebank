package cryptography

import (
	"crypto/rand"
	"io"
)

func Random12BytesNonce() ([]byte, error) {
	return randomBytes(12)
}

func randomBytes(length int) ([]byte, error) {
	nonce := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}
