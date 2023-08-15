package server

import (
	"bytes"
	"errors"
	"io"
	"log"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
	pb "github.com/oteffahi/merkle-filebank/proto"
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
	// TODO: verify signature
	validSignature, err := verifyDownloadRequestSignature(req1)
	if err != nil {
		return err
	}
	if !validSignature {
		return errors.New("Invalid challenge signature")
	}
	// TODO: verify key existence
	hasBank, err := verifyKeyHasBank(req1.Pubkey)
	if err != nil {
		return err
	}
	if !hasBank {
		return errors.New("Key does not have a bank")
	}
	// TODO: read bank descriptor from disk

	// TODO: read file from disk
	file := []byte("TEST FILE")

	// TODO: generate merkle proof
	var proof [][32]byte
	var temp [32]byte
	copy(temp[:], []byte("PROOF"))
	proof = append(proof, temp, temp)

	var linearProof []byte
	for _, hash := range proof {
		linearProof = append(linearProof, hash[:]...)
	}
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

func verifyDownloadRequestSignature(req *pb.DownloadFilesRequest) (bool, error) {
	return true, nil // TODO
}

func verifyKeyHasBank(key []byte) (bool, error) {
	return true, nil // TODO
}
