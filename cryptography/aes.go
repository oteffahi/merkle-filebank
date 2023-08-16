package cryptography

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"

	"golang.org/x/crypto/pbkdf2"
)

func EncryptData(data []byte, passphrase []byte) (ct, salt, iv []byte, err error) {
	// generate random parameters
	salt, err = randomBytes(8)
	if err != nil {
		return nil, nil, nil, err
	}
	iv, err = randomBytes(12)
	if err != nil {
		return nil, nil, nil, err
	}

	// derive key from passphrase
	derivedKey := DeriveKey(passphrase, salt)

	// create instance of cipher
	blockCipher, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, nil, nil, err
	}
	aesgcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, nil, nil, err
	}

	// encrypt
	ct = aesgcm.Seal(nil, iv, data, nil)

	return ct, salt, iv, nil
}

func DecryptData(data, aeskey, iv []byte) ([]byte, error) {
	// create instance of cipher
	blockCipher, err := aes.NewCipher(aeskey)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}

	// decrypt
	plaintext, err := aesgcm.Open(nil, iv, data, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func DeriveKey(passphrase, salt []byte) []byte {
	// derive key from passphrase with salt
	return pbkdf2.Key(passphrase, salt, 4096, 16, sha1.New)
}
