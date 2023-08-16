package client

import (
	"bytes"
	"context"
	"errors"
	"time"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
	pb "github.com/oteffahi/merkle-filebank/proto"
)

func CallAddNode(endpoint string, serverName string) error {
	nonce, err := cr.Random12BytesNonce()
	if err != nil {
		return err
	}

	conn, client, err := connectToNode(endpoint)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.AddNode(ctx, &pb.AddNodeRequest{Nonce: nonce})
	if err != nil {
		return err
	}

	if !bytes.Equal(resp.Nonce, nonce) {
		return errors.New("Invalid response message: bad nonce")
	}
	// TODO: verify signature
	validSignature, err := verifySignature(resp)
	if err != nil {
		return err
	}
	if !validSignature {
		return errors.New("Invalid response message: signature is invalid")
	}
	// TODO: write pubkey and endpoint to file
	return nil
}

func verifySignature(resp *pb.AddNodeResponse) (bool, error) {
	return true, nil // TODO
}
