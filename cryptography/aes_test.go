package cryptography

import (
	"testing"

	"golang.org/x/exp/slices"
)

func TestEncryptDecrypt(t *testing.T) {
	passphrase := []byte("testpassword")
	dataToEncrypt := []byte("DATA TO ENCRYPT")

	encryptedData, salt, iv, err := EncryptData(dataToEncrypt, passphrase)
	if err != nil {
		t.Errorf("Error occured during encrypton: %v", err)
		return
	}

	key := DeriveKey(passphrase, salt)
	decryptedData, err := DecryptData(encryptedData, key, iv)
	if err != nil {
		t.Errorf("Error occured during decryption: %v", err)
		return
	}
	if !slices.Equal(decryptedData, dataToEncrypt) {
		t.Errorf("Decrypted data different from original")
		return
	}
}
