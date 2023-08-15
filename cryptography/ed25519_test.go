package cryptography

import (
	"testing"

	pb "github.com/oteffahi/merkle-filebank/proto"
	"golang.org/x/exp/slices"
)

func TestExportImportPrivKey(t *testing.T) {
	_, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Errorf("Error occured when generating keypair: %v", err)
		return
	}

	passphrase := []byte("testpassword")
	exportedPrivKey, err := SafeExportPrivateKey(privKey, passphrase)
	if err != nil {
		t.Errorf("Error occured when exporting private key: %v", err)
		return
	}
	importedPrivKey, err := SafeImportPrivateKey(exportedPrivKey, passphrase)
	if err != nil {
		t.Errorf("Error occured when importing private key: %v", err)
		return
	}
	if !slices.Equal(importedPrivKey, privKey) {
		t.Errorf("Imported private key different from original")
		return
	}
}

func TestExportImportPubKey(t *testing.T) {
	pubKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Errorf("Error occured when generating keypair: %v", err)
		return
	}

	exportedPubKey, err := ExportPublicKey(pubKey)
	if err != nil {
		t.Errorf("Error occured when exporting public key: %v", err)
		return
	}
	importedPubKey, err := ImportPublicKey(exportedPubKey)
	if err != nil {
		t.Errorf("Error occured when importing public key: %v", err)
		return
	}
	if !slices.Equal(importedPubKey, pubKey) {
		t.Errorf("Imported public key different from original")
		return
	}
}

func TestSignVerify(t *testing.T) {
	pubKey, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Errorf("Error occured when generating keypair: %v", err)
		return
	}

	dataToSign := &pb.FileMessage{
		Seq:     1,
		Content: []byte("Example"),
	}

	signature, err := SignMessage(dataToSign, privKey)
	if err != nil {
		t.Errorf("Error occured when signing message: %v", err)
		return
	}

	if err := VerifySignature(dataToSign, pubKey, signature); err != nil {
		t.Errorf("Error occured when verifying signature: %v", err)
		return
	}
}
