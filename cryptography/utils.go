package cryptography

import (
	"crypto/rand"
	"io"
	"syscall"

	"github.com/itchyny/base58-go"
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

func Base58Encode(data []byte) (string, error) {
	encoded, err := base58.BitcoinEncoding.Encode(data)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func Base58Decode(data string) ([]byte, error) {
	encoded := []byte(data)
	return base58.BitcoinEncoding.Decode(encoded)
}

func ReadPassphrase() (string, error) {
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	return string(bytePassword), nil
}
