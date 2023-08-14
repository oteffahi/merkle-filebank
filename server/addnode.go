package server

import (
	"context"
	"log"

	pb "github.com/oteffahi/merkle-filebank/proto"
	"google.golang.org/protobuf/proto"
)

func (c *fileBankServer) AddNode(ctx context.Context, req *pb.AddNodeRequest) (*pb.AddNodeResponse, error) {
	log.Printf("Received call: AddNode")
	nonce := req.Nonce

	// TODO: get server public key

	// TODO: sign message (mockup with Marshal for now)
	m, err := proto.Marshal(req)

	if err != nil {
		return nil, err
	}

	return &pb.AddNodeResponse{
		Nonce:     nonce,
		Pubkey:    []byte("key"),
		Signature: m,
	}, nil
}
