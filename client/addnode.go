package client

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"time"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
	pb "github.com/oteffahi/merkle-filebank/proto"
	"github.com/oteffahi/merkle-filebank/storage"
)

func CallAddNode(endpoint string, bankhome string, serverName string) error {
	// verify that server does not exist locally
	if serverExists, err := storage.Client_ServerExists(bankhome, serverName); err != nil {
		return err
	} else if serverExists {
		return errors.New(fmt.Sprintf("Server %v already exist locally", serverName))
	}

	nonce, err := cr.Random12BytesNonce()
	if err != nil {
		return err
	}

	conn, client, err := connectToNode(endpoint, bankhome)
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

	// validate key as a correct ed25519 pubkey and import it
	pubKey, err := cr.ImportPublicKey(resp.Pubkey)
	if err != nil {
		return err
	}

	//verify signature
	if err := verifyAddNodeResponseSignature(resp, pubKey); err != nil {
		return err
	}

	// write pubkey and endpoint to file
	serverDescriptor := &pb.ServerDescriptor{
		PubKey: resp.Pubkey,
		Host:   endpoint,
	}
	if err := storage.Client_WriteServerDescriptor(bankhome, serverDescriptor, serverName); err != nil {
		return err
	}
	fmt.Printf("Server '%s' was successfully added to known servers\n", serverName)
	return nil
}

func verifyAddNodeResponseSignature(resp *pb.AddNodeResponse, pubKey ed25519.PublicKey) error {
	signedMessage := &pb.SignAddNodeServer{
		Nonce:  resp.Nonce,
		PubKey: resp.Pubkey,
	}
	return cr.VerifySignature(signedMessage, pubKey, resp.Signature)
}
