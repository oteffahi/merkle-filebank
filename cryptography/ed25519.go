package cryptography

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/youmark/pkcs8"
	"google.golang.org/protobuf/proto"
)

func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	return pub, priv, err
}

func SafeExportPrivateKey(key ed25519.PrivateKey, passphrase []byte) ([]byte, error) {
	pkcs8key, err := pkcs8.MarshalPrivateKey(key, passphrase, &pkcs8.Opts{Cipher: pkcs8.AES128GCM, KDFOpts: pkcs8.DefaultOpts.KDFOpts})
	if err != nil {
		return nil, err
	}

	exportedKey := pem.EncodeToMemory(&pem.Block{
		Type:  "ENCRYPTED PRIVATE KEY",
		Bytes: pkcs8key,
	})

	return exportedKey, err
}

func SafeImportPrivateKey(key []byte, passphrase []byte) (ed25519.PrivateKey, error) {
	pkcs8Key, _ := pem.Decode(key)
	importedKey, err := pkcs8.ParsePKCS8PrivateKey(pkcs8Key.Bytes, passphrase)
	if err != nil {
		return nil, err
	}
	privKey, correctType := importedKey.(ed25519.PrivateKey)
	if !correctType {
		return nil, errors.New("Imported key is not of type ed25519.PrivateKey")
	}
	return privKey, nil
}

func ExportPublicKey(key ed25519.PublicKey) ([]byte, error) {
	pkixKey, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}
	exportedKey := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pkixKey,
	})
	return exportedKey, nil
}

func ImportPublicKey(key []byte) (ed25519.PublicKey, error) {
	pkixKey, _ := pem.Decode(key)
	importedKey, err := x509.ParsePKIXPublicKey(pkixKey.Bytes)
	if err != nil {
		return nil, err
	}
	pubKey, correctType := importedKey.(ed25519.PublicKey)
	if !correctType {
		return nil, errors.New("Imported key is not of type ed25519.PublicKey")
	}
	return pubKey, nil
}

func SignMessage(m proto.Message, key ed25519.PrivateKey) ([]byte, error) {
	message, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	signature, err := key.Sign(nil, message, &ed25519.Options{Hash: crypto.Hash(0)})
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func VerifySignature(m proto.Message, key ed25519.PublicKey, signature []byte) error {
	message, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	if err := ed25519.VerifyWithOptions(key, message, signature, &ed25519.Options{Hash: crypto.Hash(0)}); err != nil {
		return err
	}

	return nil
}
