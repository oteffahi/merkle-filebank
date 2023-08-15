package client

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/oteffahi/merkle-filebank/merkle"
	pb "github.com/oteffahi/merkle-filebank/proto"
	"google.golang.org/protobuf/proto"
)

func CallDownloadFiles(endpoint string, fileNumber int) error {
	conn, client, err := connectToNode(endpoint)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 30s timeout because of file manipulations on server
	defer cancel()

	stream, err := client.DownloadFiles(ctx)
	if err != nil {
		return err
	}

	resp1, err := stream.Recv()
	if err == io.EOF {
		return errors.New("Connexion closed by server")
	}
	if err != nil {
		return err
	}

	// verify msg type is nonce
	var serverNonce []byte
	switch phase := resp1.Phase.(type) {
	case *pb.DownloadFilesResponse_Nonce:
		serverNonce = phase.Nonce
	default:
		return errors.New("Invalid message type")
	}

	// TODO: get pubkey
	pubkey := []byte("key")

	// TODO: sign nonce
	sign, err := proto.Marshal(resp1)
	if err != nil {
		return err
	}

	// generate and send request message
	if err := stream.Send(&pb.DownloadFilesRequest{
		Nonce:     serverNonce,
		Pubkey:    pubkey,
		FileNum:   int32(fileNumber),
		Signature: sign,
	}); err != nil {
		return err
	}

	resp2, err := stream.Recv()
	if err == io.EOF {
		return errors.New("Connexion closed by server")
	}
	if err != nil {
		return err
	}

	var fileAndProof *pb.FileAndProof
	switch phase := resp2.Phase.(type) {
	case *pb.DownloadFilesResponse_Fp:
		fileAndProof = phase.Fp
	default:
		return errors.New("Invalid message type")
	}

	// unlinearize merkle proof
	var serverProof [][32]byte
	if len(fileAndProof.Proof)%32 != 0 {
		return errors.New("Invalid merkle proof format")
	}
	for i := 0; i < len(fileAndProof.Proof); i += 32 {
		var buff [32]byte
		copy(buff[:], fileAndProof.Proof[i:i+32])
		serverProof = append(serverProof, buff)
	}

	// TODO: load merkleRoot
	var merkleRoot [32]byte
	copy(merkleRoot[:], []byte("ROOT"))

	// verify proof
	merkleProof := merkle.MerkleProof{
		Hashes: serverProof,
	}
	validProof, err := merkleProof.VerifyFileProof(fileAndProof.File, merkleRoot)
	if err != nil {
		return err
	}
	if validProof { //TODO: flip boolean when verification is implemented
		return errors.New("Invalid merkle proof")
	}

	// TODO: write file to disk

	return nil
}
