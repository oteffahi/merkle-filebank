package server

import (
	"context"
	"errors"
	"log"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
	pb "github.com/oteffahi/merkle-filebank/proto"
)

func (c *fileBankServer) AddNode(ctx context.Context, req *pb.AddNodeRequest) (*pb.AddNodeResponse, error) {
	log.Printf("Received call: AddNode")
	nonce := req.Nonce

	// sign message
	if ServerKeys.privKey == nil || ServerKeys.pubKey == nil {
		log.Println("Error occured while precessing call for AddNode: Keypair not loaded.")
		return nil, errors.New("Internal server error")
	}

	// export pubkey
	exportedPubKey, err := cr.ExportPublicKey(ServerKeys.pubKey)
	if err != nil {
		log.Printf("Error occured while precessing call for AddNode: %v\n", err)
		return nil, errors.New("Internal server error")
	}

	messageToSign := &pb.SignAddNodeServer{
		Nonce:  nonce,
		PubKey: exportedPubKey,
	}
	signature, err := cr.SignMessage(messageToSign, ServerKeys.privKey)
	if err != nil {
		log.Printf("Error occured while precessing call for AddNode: %v\n", err)
		return nil, errors.New("Internal server error")
	}

	return &pb.AddNodeResponse{
		Nonce:     nonce,
		Pubkey:    exportedPubKey,
		Signature: signature,
	}, nil
}
