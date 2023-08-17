package cryptography

import (
	"crypto/rand"
	"io"
	"syscall"

	"github.com/akamensky/base58"
	"golang.org/x/term"
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

func Base58Encode(data []byte) string {
	encoded := base58.Encode(data)
	return string(encoded)
}

func Base58Decode(data string) ([]byte, error) {
	return base58.Decode(data)
}

func ReadPassphrase() (string, error) {
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	return string(bytePassword), nil
}
