package server

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"log"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
	"github.com/oteffahi/merkle-filebank/merkle"
	pb "github.com/oteffahi/merkle-filebank/proto"
	"github.com/oteffahi/merkle-filebank/storage"
)

func (c *fileBankServer) DownloadFiles(stream pb.FileBankService_DownloadFilesServer) error {
	log.Printf("Received call: DownloadFiles")
	serverNonce, err := cr.Random12BytesNonce()
	if err != nil {
		return err
	}

	if err := stream.Send(&pb.DownloadFilesResponse{
		Phase: &pb.DownloadFilesResponse_Nonce{
			Nonce: serverNonce,
		},
	}); err != nil {
		return err
	}

	// verify nonce
	req1, err := stream.Recv()
	if err == io.EOF {
		return errors.New("Connexion closed by client")
	}
	if err != nil {
		return err
	}

	// verify nonce matches
	if !bytes.Equal(req1.Nonce, serverNonce) {
		return errors.New("Invalid challenge response nonce")
	}

	// verify bank existence
	if exists, err := verifyBankExistenceFromAddress(req1.PubKeyAddr); err != nil {
		return err
	} else if !exists {
		return errors.New("Bank does not exist")
	}
	// read bank descriptor from disk
	bankDescriptor, err := storage.Server_ReadBankDescriptor(bankhome, req1.PubKeyAddr)
	if err != nil {
		return err
	}

	// import bank public key
	pubKey, err := cr.ImportPublicKey(bankDescriptor.PubKey)
	if err != nil {
		return err
	}

	// verify signature
	if err := verifyDownloadRequestSignature(req1, pubKey); err != nil {
		return err
	}

	if req1.FileNum < 1 || req1.FileNum > bankDescriptor.Nbfiles {
		return errors.New(fmt.Sprintf("No file identified by %v. Bank %v has files between 1-%v", req1.FileNum, req1.PubKeyAddr, bankDescriptor.Nbfiles))
	}

	// read file from disk
	file, err := storage.Server_ReadFileFromBank(bankhome, req1.PubKeyAddr, int(req1.FileNum))
	if err != nil {
		return err
	}

	// convert bankDescriptor.MerkleHashes from slice of slices to slice of arrays
	merkleHashes := [][32]byte{}
	for _, hash := range bankDescriptor.MerkleHashes {
		merkleHashes = append(merkleHashes, [32]byte(hash))
	}
	// load merkle tree
	merkleTree := merkle.MerkleTree{
		Hashes: merkleHashes,
	}
	// generate proof
	merkleProof, err := merkleTree.GenerateProofForFile(file)
	if err != nil {
		return err
	}

	// linearize proof to fit in one message
	var linearProof []byte
	for _, hash := range merkleProof.Hashes {
		linearProof = append(linearProof, hash[:]...)
	}

	// send file and proof
	resp := &pb.DownloadFilesResponse{
		Phase: &pb.DownloadFilesResponse_Fp{
			Fp: &pb.FileAndProof{
				Proof: linearProof,
				File:  file,
			},
		},
	}
	if err := stream.Send(resp); err != nil {
		return err
	}

	return nil
}

func verifyDownloadRequestSignature(req *pb.DownloadFilesRequest, pubKey ed25519.PublicKey) error {
	clientSignedMsg := &pb.SignDownloadRequestClient{
		Nonce:      req.Nonce,
		PubKeyAddr: req.PubKeyAddr,
		FileNum:    req.FileNum,
	}
	return cr.VerifySignature(clientSignedMsg, pubKey, req.Signature)
}

func verifyBankExistenceFromAddress(keyHashB58 string) (bool, error) {
	if exists, err := storage.Server_BankExists(bankhome, keyHashB58); err != nil {
		return false, err
	} else if exists {
		return true, nil
	}
	return false, nil
}
