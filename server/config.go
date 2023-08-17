package server

import (
	"crypto/ed25519"
	"errors"
)

type ServerKeyPair struct {
	privKey ed25519.PrivateKey
	pubKey  ed25519.PublicKey
}

var ServerKeys *ServerKeyPair // global instance
var bankhome string           // global path

func LoadKeyPair(privKey ed25519.PrivateKey) error {
	if ServerKeys == nil {
		pubKey, ok := privKey.Public().(ed25519.PublicKey)
		if !ok {
			return errors.New("Invalid public key format")
		}

		ServerKeys = &ServerKeyPair{
			privKey: privKey,
			pubKey:  pubKey,
		}
		return nil
	}
	return errors.New("Key pair already loaded")
}

func SetBankHome(path string) error {
	if bankhome == "" {
		bankhome = path
		return nil
	} else {
		return errors.New("Cannot overwrite home path while server is running")
	}
}
