package cryptography

import (
	"crypto/rand"
	"io"
)

func Random12BytesNonce() ([]byte, error) {
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

func ValidateNonce(received []byte, sent []byte) bool {
	if len(received) != len(sent) {
		return false
	}
	for i := 0; i < len(received); i++ {
		if received[i] != sent[i] {
			return false
		}
	}
	return true
}
